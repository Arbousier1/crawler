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
)

const (
	BaseURL       = "https://xiao-momi.github.io/craft-engine-wiki/"
	OutDir        = "dist"
	FinalPDF      = "Wiki_Full_Dump.pdf"
	MaxConcurrent = 4
)

const CleanScript = `
	document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe, .theme-container > .navbar').forEach(e => e.remove());
	document.querySelectorAll('details').forEach(e => e.open = true);
	document.body.style.padding = '0px';
	document.body.style.margin = '20px';
	document.body.style.backgroundColor = 'white';
	document.querySelectorAll('.sidebar-mask').forEach(e => e.remove());
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
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🕷️ 启动深度爬虫 (Breadth-First Search)...")
	urls := crawlAllPages(allocCtx)
	
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 最终捕获: %d 个唯一页面\n", len(uniqueUrls))

	if len(uniqueUrls) == 0 {
		log.Fatal("❌ 未找到任何页面")
	}

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
	fmt.Printf("🏆 完成！耗时: %s | 文件: %s\n", time.Since(start), FinalPDF)
	os.RemoveAll(OutDir)
}

func crawlAllPages(rootCtx context.Context) []string {
	ctx, cancel := chromedp.NewContext(rootCtx)
	defer cancel()

	queue := []string{BaseURL}
	seen := make(map[string]bool)
	seen[BaseURL] = true
	var results []string

	for len(queue) > 0 {
		currentURL := queue[0]
		queue = queue[1:]
		
		results = append(results, currentURL)
		fmt.Printf("🔍 扫描中 [%d Found]: %s\n", len(results), currentURL)

		newLinks := extractLinks(ctx, currentURL)
		
		for _, link := range newLinks {
			u, err := url.Parse(link)
			if err != nil { continue }
			u.Fragment = ""
			normalizedLink := strings.TrimSuffix(u.String(), "/")

			if strings.HasPrefix(normalizedLink, BaseURL) && !seen[normalizedLink] {
				seen[normalizedLink] = true
				queue = append(queue, normalizedLink)
			}
		}
	}
	return results
}

func extractLinks(ctx context.Context, targetURL string) []string {
	tCtx, cancel := context.WithTimeout(ctx, 20*time.Second) // 增加超时防止卡死
	defer cancel()

	var res []string
	err := chromedp.Run(tCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a => a.href)`, &res),
	)
	
	if err != nil {
		fmt.Printf("⚠️ 无法扫描页面: %s (%v)\n", targetURL, err)
		return []string{}
	}
	return res
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.otf", "*.mp4", "*google-analytics*",
	}))

	for t := range tasks {
		var buf []byte // 这里声明了外部变量
		tCtx, tCancel := context.WithTimeout(ctx, 45*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1500*time.Millisecond),
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				// 【修复点】：使用 = 而不是 :=，并显式声明 err
				// 这样数据才会写入到外部的 buf 中
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(false).
					WithPaperWidth(8.27).WithPaperHeight(11.69).
					WithMarginTop(0.3).WithMarginBottom(0.3).
					WithMarginLeft(0.3).WithMarginRight(0.3).
					Do(ctx)
				return err
			}),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 渲染失败: %s\n", t.URL)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("📄 [%d] 保存: %s\n", t.ID, t.URL)
	}
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	fmt.Println("📚 正在合并 PDF...")
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}
	// 传入 nil 使用默认配置
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, nil); err != nil {
		log.Printf("Merge error: %v", err)
	}
}

func uniqueAndSort(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	sort.Strings(list)
	return list
}
