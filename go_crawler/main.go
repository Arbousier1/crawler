package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
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
	SitemapURL    = "https://xiao-momi.github.io/craft-engine-wiki/sitemap.xml"
	OutDir        = "dist"
	FinalPDF      = "Wiki_Full_Coverage.pdf"
	MaxConcurrent = 4
)

// DOM 净化脚本
const CleanScript = `
	document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe').forEach(e => e.remove());
	document.querySelectorAll('details').forEach(e => e.open = true);
	document.body.style.padding = '0px';
	document.body.style.margin = '20px';
	document.body.style.backgroundColor = 'white';
`

// 暴力展开菜单脚本 (用于保底爬取)
const ExpandScript = `
	// 尝试点击所有可能的展开按钮
	document.querySelectorAll('.toggle, .arrow, button[aria-expanded="false"]').forEach(b => b.click());
	// 滚动到底部触发懒加载
	window.scrollTo(0, document.body.scrollHeight);
`

type Task struct {
	ID  int
	URL string
}

type Result struct {
	ID   int
	Path string
}

// Sitemap 结构定义
type UrlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Urls    []Url    `xml:"url"`
}
type Url struct {
	Loc string `xml:"loc"`
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
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// --- 核心改动：双重链接发现机制 ---
	fmt.Println("🔍 正在尝试获取 Sitemap (最全方式)...")
	urls := getLinksFromSitemap()

	if len(urls) == 0 {
		fmt.Println("⚠️ Sitemap 获取失败，切换到【暴力爬虫模式】...")
		urls = scanLinksAggressive(allocCtx)
	}
	
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 最终确认页面数量: %d 个 (覆盖率提升)\n", len(uniqueUrls))

	// 并发渲染
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
	fmt.Printf("🏆 任务完成！耗时: %s | 文件: %s\n", time.Since(start), FinalPDF)
	os.RemoveAll(OutDir)
}

// 策略 1: 从 Sitemap 获取 (100% 准确)
func getLinksFromSitemap() []string {
	resp, err := http.Get(SitemapURL)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var urlSet UrlSet
	if err := xml.Unmarshal(data, &urlSet); err != nil {
		return nil
	}

	var links []string
	for _, u := range urlSet.Urls {
		// 过滤掉非 HTML 文件 (如 PDF, 图片)
		if strings.HasSuffix(u.Loc, ".html") || strings.HasSuffix(u.Loc, "/") {
			links = append(links, u.Loc)
		}
	}
	return links
}

// 策略 2: 暴力爬取 (保底方案)
func scanLinksAggressive(ctx context.Context) []string {
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
		tCtx, tCancel := context.WithTimeout(ctx, 30*time.Second) // 增加扫描时间
		chromedp.Run(tCtx,
			chromedp.Navigate(curr),
			chromedp.WaitReady("body"),
			// 关键：暴力展开菜单 + 滚动页面
			chromedp.Evaluate(ExpandScript, nil),
			chromedp.Sleep(1*time.Second), // 等待展开动画
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

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.mp4", "*google-analytics*",
	}))

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1500*time.Millisecond), // 稍微多等一下图片
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
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
			fmt.Printf("⚠️ Skip: %s\n", t.URL)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("📄 [%d/%d] OK: %s\n", t.ID+1, cap(tasks), t.URL)
	}
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	fmt.Println("📚 Merging PDFs...")
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, nil); err != nil {
		fmt.Println("Merge error:", err)
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
