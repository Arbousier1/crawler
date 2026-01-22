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
)

// 针对该站点的 CSS 净化脚本：保留主体，剔除所有干扰项
const CleanScript = `
	(function() {
		const selectors = [
			'.navbar', 
			'.theme-doc-sidebar-container', 
			'.table-of-contents', 
			'.pagination-nav', 
			'footer',
			'.theme-doc-footer-edit-meta-row',
			'nav[aria-label="Breadcrumbs"]', // 移除面包屑导航
			'.admonition' // 可选：如果不需要警告框可以移除，建议保留
		];
		selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));

		// 移除 Docusaurus 默认的最大宽度限制，防止 PDF 左右留白过多
		const containers = ['.main-wrapper', '.theme-doc-item-container', '.container'];
		containers.forEach(s => {
			const el = document.querySelector(s);
			if(el) {
				el.style.maxWidth = 'none';
				el.style.padding = '10px';
				el.style.margin = '0';
			}
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
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 1. 深度扫描
	fmt.Println("🔍 正在扫描 Wiki 全站架构 (深度模式)...")
	urls := scanLinksDeep(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	
	if len(uniqueUrls) <= 1 {
		fmt.Printf("⚠️ 警告：仅发现 %d 个页面，请检查网络或 BaseURL 是否正确。\n", len(uniqueUrls))
	} else {
		fmt.Printf("✅ 扫描完成！发现 %d 个文档页面，开始并发渲染...\n", len(uniqueUrls))
	}

	// 2. 并发渲染
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

	// 3. 排序与合并
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	mergePDFs(results)
	fmt.Printf("\n🏆 全部完成！\n⏱️ 耗时: %s\n📄 输出: %s\n", time.Since(start), FinalPDF)
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
			network.SetBlockedURLs([]string{"*.woff*", "*.ttf", "*analytics*"}),
			chromedp.Navigate(t.URL),
			// 确保文章正文加载完成
			chromedp.WaitReady("article"), 
			chromedp.Sleep(2*time.Second), // 额外缓冲让图片加载
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

// 深度扫描逻辑：会递归寻找所有属于 /docs/ 路径的链接
func scanLinksDeep(ctx context.Context) []string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	
	visited := make(map[string]bool)
	var links []string
	
	// 入口队列：从首页开始
	queue := []string{BaseURL}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		cleanURL := strings.TrimSuffix(curr, "/")
		if visited[cleanURL] { continue }
		visited[cleanURL] = true

		fmt.Printf("🔗 正在探测: %s\n", cleanURL)

		var res []string
		tCtx, tCancel := context.WithTimeout(ctx, 20*time.Second)
		err := chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			// 等待侧边栏或主要链接加载
			chromedp.WaitReady("a[href]"),
			// 获取所有站内链接
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('a[href]'))
					.map(a => a.href)
					.filter(href => href.startsWith(window.location.origin + "/customcrops"))
			`, &res),
		)
		tCancel()

		if err != nil { continue }

		for _, l := range res {
			u, _ := url.Parse(l)
			u.Fragment = "" // 移除 #锚点
			u.RawQuery = "" // 移除查询参数
			full := strings.TrimSuffix(u.String(), "/")

			// 只要是该站点的文档链接且未访问过，就加入队列
			if !visited[full] {
				// 记录有效的文档链接
				if strings.Contains(full, "/docs/") || full == BaseURL {
					links = append(links, full)
					queue = append(queue, full) // 递归探测
				}
			}
		}
	}
	return links
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}
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
		if !m[v] {
			m[v] = true
			list = append(list, v)
		}
	}
	sort.Strings(list)
	return list
}
