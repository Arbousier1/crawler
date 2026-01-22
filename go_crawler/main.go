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
    // 移除 model 引用，避免版本兼容性问题
)

const (
	BaseURL       = "https://xiao-momi.github.io/craft-engine-wiki/"
	OutDir        = "dist"
	FinalPDF      = "Wiki_Multimodal_AI.pdf"
	MaxConcurrent = 4 
)

// DOM 净化脚本：保留图片，删除无关元素
const CleanScript = `
	document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe').forEach(e => e.remove());
	document.querySelectorAll('details').forEach(e => e.open = true);
	document.body.style.padding = '0px';
	document.body.style.margin = '20px';
	document.body.style.backgroundColor = 'white';
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

	// 1. 启动浏览器配置
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true), 
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 2. 扫描链接
	fmt.Println("⚡ 正在扫描全站链接...")
	urls := scanLinks(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 扫描完成: %d 个唯一页面，开始并发渲染...\n", len(uniqueUrls))

	// 3. 并发流水线
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

	// 4. 收集结果
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	// 5. 极速合并
	mergePDFs(results)

	fmt.Printf("🏆 任务完成！耗时: %s | 文件: %s\n", time.Since(start), FinalPDF)
	os.RemoveAll(OutDir)
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	// 网络拦截：只拦截绝对不需要的资源
	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.otf",
		"*.mp4", "*.webm", "*.mp3",
		"*google-analytics*", "*hm.baidu*",
	}))

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1*time.Second), // 等待图片懒加载
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(false). // 不打印背景色，保留内容图
					WithPaperWidth(8.27).
					WithPaperHeight(11.69).
					WithMarginTop(0.3). WithMarginBottom(0.3).
					WithMarginLeft(0.3). WithMarginRight(0.3).
					Do(ctx)
				return err
			}),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 失败 [%d]: %s (%v)\n", t.ID, t.URL, err)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("🖼️ [%d/%d] 完成: %s\n", t.ID+1, cap(tasks), t.URL)
	}
}

func scanLinks(ctx context.Context) []string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	
	var links []string
	toVisit := []string{BaseURL}
	visited := make(map[string]bool)
	
	for len(toVisit) > 0 {
		curr := toVisit[0]
		toVisit = toVisit[1:]
		if visited[curr] { continue }
		visited[curr] = true
		links = append(links, curr)

		var res []string
		tCtx, tCancel := context.WithTimeout(ctx, 15*time.Second)
		chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &res),
		)
		tCancel()

		for _, l := range res {
			u, err := url.Parse(l)
			if err != nil { continue }
			u.Fragment = ""
			full := strings.TrimSuffix(u.String(), "/")
			if strings.HasPrefix(full, BaseURL) && !visited[full] {
				toVisit = append(toVisit, full)
			}
		}
	}
	return links
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	fmt.Println("📚 正在合并 PDF (使用默认配置)...")
	
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}

	// 修复点：直接传入 nil 作为配置
	// pdfcpu 会自动使用 DefaultConfig，这避免了版本不兼容导致的 undefined 错误
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, nil); err != nil {
		log.Fatal(err)
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
