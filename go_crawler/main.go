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
	BaseURL       = "https://momi.gtemc.cn/customcrops"
	OutDir        = "dist"
	FinalPDF      = "MOMI_CustomCrops_Wiki.pdf"
	MaxConcurrent = 4
	// 模拟真实浏览器
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

const CleanScript = `
	(function() {
		const selectors = ['.navbar', '.theme-doc-sidebar-container', '.table-of-contents', '.pagination-nav', 'footer', '.theme-doc-footer-edit-meta-row', 'nav[aria-label="Breadcrumbs"]'];
		selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));
		const containers = ['.main-wrapper', '.theme-doc-item-container', '.container'];
		containers.forEach(s => {
			const el = document.querySelector(s);
			if(el) { el.style.maxWidth = 'none'; el.style.padding = '10px'; el.style.margin = '0'; }
		});
		document.querySelectorAll('details').forEach(e => e.open = true);
		document.body.style.backgroundColor = 'white';
	})();
`

type Task struct {
	ID  int
	URL string
}

type Result struct {
	ID   int
	Path string
}

func main() {
	start := time.Now()
	os.RemoveAll(OutDir)
	os.MkdirAll(OutDir, 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(UserAgent), // 设置伪装 UA
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🔍 正在扫描 Wiki 全站架构 (深度模式)...")
	urls := scanLinksDeep(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	
	if len(uniqueUrls) == 0 {
		fmt.Println("❌ 错误：未能获取任何有效链接。程序退出。")
		os.Exit(1)
	}

	fmt.Printf("✅ 发现 %d 个有效页面，开始并发生成...\n", len(uniqueUrls))

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
	fmt.Printf("\n🏆 全部完成！耗时: %s | 输出: %s\n", time.Since(start), FinalPDF)
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 90*time.Second)
		err := chromedp.Run(tCtx,
			network.Enable(),
			chromedp.Navigate(t.URL),
			chromedp.WaitVisible("main", chromedp.ByQuery), // 改为等待 main 标签可见
			chromedp.Sleep(2*time.Second),
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().WithPrintBackground(true).WithPaperWidth(8.27).WithPaperHeight(11.69).Do(ctx)
				return err
			}),
		)
		tCancel()

		if err != nil {
			fmt.Printf("❌ [%d] 失败: %s\n", t.ID, t.URL)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("📄 [%d] 已生成: %s\n", t.ID, t.URL)
	}
}

func scanLinksDeep(ctx context.Context) []string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	
	visited := make(map[string]bool)
	var links []string
	queue := []string{BaseURL}

	// 统一域名识别
	targetHost := "momi.gtemc.cn"

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
		tCtx, tCancel := context.WithTimeout(ctx, 25*time.Second)
		err := chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.Sleep(3*time.Second), // 强制等待 JS 渲染侧边栏
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('a[href]'))
					.map(a => a.href)
			`, &res),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 探测页面出错: %v\n", err)
			continue
		}

		foundNew := 0
		for _, l := range res {
			parsed, err := url.Parse(l)
			if err != nil { continue }

			// 只要是该域名下的链接，且不是静态文件
			if parsed.Host == targetHost && 
			   strings.HasPrefix(parsed.Path, "/customcrops") && 
			   !strings.HasSuffix(parsed.Path, ".png") && 
			   !strings.HasSuffix(parsed.Path, ".jpg") {
				
				parsed.Fragment = ""
				parsed.RawQuery = ""
				full := strings.TrimSuffix(parsed.String(), "/")

				if !visited[full] {
					// 只有 /docs/ 路径或者 BaseURL 才记录为待打印页面
					if strings.Contains(full, "/docs/") || full == BaseURL {
						links = append(links, full)
					}
					queue = append(queue, full)
					foundNew++
				}
			}
		}
		if foundNew > 0 {
			fmt.Printf("   ├─ 发现 %d 个新链接...\n", foundNew)
		}
	}
	return links
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	var inFiles []string
	for _, r := range results { inFiles = append(inFiles, r.Path) }
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed 
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, conf); err != nil {
		log.Fatalf("❌ 合并 PDF 失败: %v", err)
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
