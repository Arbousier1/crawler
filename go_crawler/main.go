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
	// 更新为 GitBook 地址
	BaseURL       = "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops"
	OutDir        = "dist"
	FinalPDF      = "MOMI_CustomCrops_Wiki.pdf"
	MaxConcurrent = 3 // GitBook 比较稳定，可以稍微提高并发
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

// GitBook 专属净化脚本
const CleanScript = `
	(function() {
		// 移除侧边栏、顶部导航头、右侧反馈按钮等
		const selectors = [
			'header', 
			'nav', 
			'[role="navigation"]', 
			'#feedback-buoy', 
			'footer',
			'.css-175oi2r.r-13awgt0.r-1777fci' // 常见的 GitBook 遮罩/页脚类
		];
		selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));

		// 强制内容区域占满全屏
		const main = document.querySelector('main');
		if(main) {
			main.style.width = '100%';
			main.style.maxWidth = 'none';
			main.style.margin = '0';
			main.style.padding = '20px';
		}

		// 移除可能存在的最大宽度限制
		document.querySelectorAll('div').forEach(div => {
			if (window.getComputedStyle(div).maxWidth !== 'none') {
				div.style.maxWidth = 'none';
			}
		});

		document.body.style.backgroundColor = 'white';
	})();
`

func main() {
	start := time.Now()
	os.RemoveAll(OutDir)
	os.MkdirAll(OutDir, 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(UserAgent),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🔍 正在扫描 GitBook 目录结构...")
	uniqueUrls := scanLinksDeep(allocCtx)
	
	if len(uniqueUrls) == 0 {
		fmt.Println("❌ 错误：未能从 GitBook 获取任何有效链接。")
		os.Exit(1)
	}

	fmt.Printf("✅ 发现 %d 个文档页面，开始生成 PDF...\n", len(uniqueUrls))

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

	mergePDFs(results)
	fmt.Printf("\n🏆 完成！耗时: %s | 输出: %s\n", time.Since(start), FinalPDF)
}

func scanLinksDeep(ctx context.Context, ) []string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	
	visited := make(map[string]bool)
	var links []string
	queue := []string{BaseURL}
	targetHost := "mo-mi.gitbook.io"
	basePath := "/xiaomomi-plugins/customcrops"

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		u, _ := url.Parse(curr)
		cleanURL := u.Scheme + "://" + u.Host + u.Path
		cleanURL = strings.TrimSuffix(cleanURL, "/")

		if visited[cleanURL] { continue }
		visited[cleanURL] = true

		fmt.Printf("🔗 正在探测: %s\n", cleanURL)

		var res []string
		tCtx, tCancel := context.WithTimeout(ctx, 40*time.Second)
		err := chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.WaitReady("body"),
			chromedp.Sleep(3*time.Second), // 等待 GitBook 加载侧边栏
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('a[href]'))
					.map(a => a.href)
			`, &res),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 跳过页面: %v\n", err)
			continue
		}

		for _, l := range res {
			parsed, err := url.Parse(l)
			if err != nil { continue }

			// 检查是否属于同一个 GitBook 项目
			if parsed.Host == targetHost && strings.HasPrefix(parsed.Path, basePath) {
				parsed.Fragment = ""
				parsed.RawQuery = ""
				full := strings.TrimSuffix(parsed.String(), "/")

				if !visited[full] {
					links = append(links, full)
					queue = append(queue, full) // 递归抓取侧边栏里的所有链接
				}
			}
		}
	}
	return uniqueAndSort(links)
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
		err := chromedp.Run(tCtx,
			network.Enable(),
			chromedp.Navigate(t.URL),
			chromedp.WaitVisible("main", chromedp.ByQuery),
			chromedp.Sleep(2*time.Second),
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

		if err != nil {
			fmt.Printf("❌ [%d] 渲染失败: %s\n", t.ID, t.URL)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("📄 [%d] 已生成: %s\n", t.ID, t.URL)
	}
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	var inFiles []string
	for _, r := range results { inFiles = append(inFiles, r.Path) }
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed 
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, conf); err != nil {
		log.Printf("❌ 合并出错: %v", err)
	}
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
