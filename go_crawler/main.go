package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

const (
	BaseURL       = "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops"
	OutDir        = "dist"
	FinalPDF      = "MOMI_CustomCrops_Wiki.pdf"
	// 稳定性核心：在 GitHub Actions 中建议设为 1 或 2，防止浏览器崩溃
	MaxConcurrent = 1 
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

func main() {
	start := time.Now()
	os.RemoveAll(OutDir)
	os.MkdirAll(OutDir, 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.UserAgent(UserAgent),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🔍 正在扫描 GitBook 目录结构...")
	uniqueUrls := scanLinksDeep(allocCtx)
	
	if len(uniqueUrls) == 0 {
		fmt.Println("❌ 错误：未能获取有效链接。")
		os.Exit(1)
	}

	fmt.Printf("✅ 发现 %d 个文档页面，开始串行稳定渲染...\n", len(uniqueUrls))

	taskChan := make(chan Task, len(uniqueUrls))
	resChan := make(chan Result, len(uniqueUrls))
	var wg sync.WaitGroup

	for i := 0; i < MaxConcurrent; i++ {
		wg.Add(1)
		go worker(allocCtx, taskChan, resChan, &wg)
	}

	for i, u := range uniqueUrls {
		taskChan <- Task{ID: i, URL: u}
	}
	close(taskChan)
	wg.Wait()
	close(resChan)

	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	if len(results) == 0 {
		fmt.Println("❌ 错误：所有页面渲染均失败，无法合并。")
		os.Exit(1)
	}

	mergePDFs(results)
	fmt.Printf("\n🏆 完成！成功渲染 %d/%d 页 | 耗时: %s\n", len(results), len(uniqueUrls), time.Since(start))
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for t := range tasks {
		success := false
		var buf []byte
		
		// 自动重试机制：最多尝试 3 次
		for attempt := 1; attempt <= 3; attempt++ {
			if attempt > 1 {
				fmt.Printf("🔄 [%d] 正在进行第 %d 次重试...\n", t.ID, attempt)
				time.Sleep(2 * time.Second)
			}

			// 为每次渲染创建完全独立的 Context，防止互相干扰
			ctx, cancel := chromedp.NewContext(parentCtx)
			tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
			
			err := chromedp.Run(tCtx,
				network.Enable(),
				chromedp.Navigate(t.URL),
				// GitBook 加载较慢，等待 body 出现即可
				chromedp.WaitReady("body"),
				chromedp.Sleep(5*time.Second), // 留出足够时间给动态内容
				chromedp.Evaluate(CleanScript, nil),
				chromedp.ActionFunc(func(ctx context.Context) error {
					var err error
					buf, _, err = page.PrintToPDF().
						WithPrintBackground(true).
						WithPaperWidth(8.27).
						WithPaperHeight(11.69).
						Do(ctx)
					return err
				}),
			)
			tCancel()
			cancel() // 渲染完立即释放浏览器 Tab 内存

			if err == nil {
				success = true
				break
			}
			fmt.Printf("⚠️ [%d] 尝试 %d 失败: %v\n", t.ID, attempt, err)
		}

		if success {
			path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
			os.WriteFile(path, buf, 0644)
			results <- Result{ID: t.ID, Path: path}
			fmt.Printf("📄 [%d] 渲染成功: %s\n", t.ID, t.URL)
		} else {
			fmt.Printf("❌ [%d] 最终渲染失败: %s\n", t.ID, t.URL)
		}
	}
}

func scanLinksDeep(ctx context.Context) []string {
	// 扫描使用独立的 Context
	sCtx, sCancel := chromedp.NewContext(ctx)
	defer sCancel()
	
	visited := make(map[string]bool)
	var links []string
	queue := []string{BaseURL}
	targetHost := "mo-mi.gitbook.io"

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		u, _ := url.Parse(curr)
		cleanURL := u.Scheme + "://" + u.Host + u.Path
		cleanURL = strings.TrimSuffix(cleanURL, "/")

		if visited[cleanURL] { continue }
		visited[cleanURL] = true

		fmt.Printf("🔗 正在扫描: %s\n", cleanURL)

		var res []string
		tCtx, tCancel := context.WithTimeout(sCtx, 30*time.Second)
		err := chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.WaitReady("body"),
			chromedp.Sleep(3*time.Second),
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('a[href]'))
					.map(a => a.href)
			`, &res),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 扫描页面出错 (跳过): %v\n", err)
			continue
		}

		for _, l := range res {
			parsed, err := url.Parse(l)
			if err != nil { continue }

			if parsed.Host == targetHost && strings.Contains(parsed.Path, "customcrops") {
				parsed.Fragment = ""
				parsed.RawQuery = ""
				full := strings.TrimSuffix(parsed.String(), "/")

				if !visited[full] {
					links = append(links, full)
					queue = append(queue, full)
				}
			}
		}
	}
	return uniqueAndSort(links)
}

const CleanScript = `
	(function() {
		// 彻底移除干扰元素
		const selectors = ['header', 'nav', '[role="navigation"]', '#feedback-buoy', 'footer', 'iframe'];
		selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));

		const main = document.querySelector('main');
		if(main) {
			main.style.width = '100%';
			main.style.maxWidth = 'none';
			main.style.margin = '0';
			main.style.padding = '30px';
		}
		document.body.style.backgroundColor = 'white';
	})();
`

func mergePDFs(results []Result) {
	var inFiles []string
	for _, r := range results { inFiles = append(inFiles, r.Path) }
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed 
	api.MergeCreateFile(inFiles, FinalPDF, false, conf)
}

func uniqueAndSort(slice []string) []string {
	m := make(map[string]bool)
	var list []string
	for _, v := range slice {
		if !m[v] && v != "" {
			m[v] = true
			list = append(list, v)
		}
	}
	sort.Strings(list)
	return list
}

type Task struct {
	ID  int
	URL string
}

type Result struct {
	ID   int
	Path string
}
