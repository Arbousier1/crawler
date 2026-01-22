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
	MaxConcurrent = 4 // 保持稳定
)

// 净化脚本：保留图片，删除导航
const CleanScript = `
	document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe, .theme-container > .navbar').forEach(e => e.remove());
	document.querySelectorAll('details').forEach(e => e.open = true);
	document.body.style.padding = '0px';
	document.body.style.margin = '20px';
	document.body.style.backgroundColor = 'white';
	// 尝试移除 VuePress/VitePress 的遮罩
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

	// 1. 浏览器配置 (关键：设置大窗口)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		// 【关键修复】强制 1920x1080，防止侧边栏被折叠
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 2. 深度递归扫描
	fmt.Println("🕷️ 启动深度爬虫 (Breadth-First Search)...")
	urls := crawlAllPages(allocCtx)
	
	// 再次去重，确保万无一失
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 最终捕获: %d 个唯一页面 (准备渲染)\n", len(uniqueUrls))

	if len(uniqueUrls) == 0 {
		log.Fatal("❌ 未找到任何页面，请检查 BaseURL 是否可访问")
	}

	// 3. 并发渲染
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

	// 4. 合并
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	mergePDFs(results)
	fmt.Printf("🏆 完成！耗时: %s | 文件: %s\n", time.Since(start), FinalPDF)
	os.RemoveAll(OutDir)
}

// crawlAllPages 实现了真正的 BFS (广度优先搜索)
func crawlAllPages(rootCtx context.Context) []string {
	// 创建一个独立的 browser context 用于爬取
	ctx, cancel := chromedp.NewContext(rootCtx)
	defer cancel()

	// 待爬队列
	queue := []string{BaseURL}
	// 已发现集合 (用于去重)
	seen := make(map[string]bool)
	seen[BaseURL] = true
	// 结果列表
	var results []string

	// 限制最大深度防止死循环 (Wiki一般不超过5层，但这里按数量限制更安全)
	// 或者只要队列不空就一直爬
	for len(queue) > 0 {
		// 取出队首
		currentURL := queue[0]
		queue = queue[1:]
		
		results = append(results, currentURL)
		fmt.Printf("🔍 扫描中 [%d Found]: %s\n", len(results), currentURL)

		// 提取该页面上的所有新链接
		newLinks := extractLinks(ctx, currentURL)
		
		for _, link := range newLinks {
			// 规范化链接：去掉锚点，去掉尾部斜杠
			u, err := url.Parse(link)
			if err != nil { continue }
			u.Fragment = ""
			normalizedLink := strings.TrimSuffix(u.String(), "/")

			// 必须是站内链接，且未被发现过
			if strings.HasPrefix(normalizedLink, BaseURL) && !seen[normalizedLink] {
				seen[normalizedLink] = true
				queue = append(queue, normalizedLink)
			}
		}
	}
	return results
}

func extractLinks(ctx context.Context, targetURL string) []string {
	// 设置超时，防止某个页面卡死
	tCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var res []string
	err := chromedp.Run(tCtx,
		chromedp.Navigate(targetURL),
		// 等待侧边栏加载 (VuePress 常见的选择器)
		chromedp.WaitReady("body"),
		// 稍微睡一下，等 JS 渲染侧边栏
		chromedp.Sleep(1*time.Second),
		// 抓取所有链接
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a => a.href)`, &res),
	)
	
	if err != nil {
		// 超时或出错也不要 panic，直接返回空，继续下一个
		fmt.Printf("⚠️ 无法扫描页面: %s (%v)\n", targetURL, err)
		return []string{}
	}
	return res
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	// 拦截无用资源 (只拦截字体和视频，保留图片)
	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.otf", "*.mp4", "*google-analytics*",
	}))

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 45*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1500*time.Millisecond), // 等图片
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				buf, _, err := page.PrintToPDF().
					WithPrintBackground(false). // 不打印背景
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
