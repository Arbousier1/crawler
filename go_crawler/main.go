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
	MaxConcurrent = 4 // GitHub Actions 建议值
)

// 针对 momi.gtemc.cn 的净化脚本
const CleanScript = `
	// 1. 移除导航栏、侧边栏、右侧目录、底部导航、页脚
	const selectors = [
		'.navbar', 
		'.theme-doc-sidebar-container', 
		'.table-of-contents', 
		'.pagination-nav', 
		'footer',
		'.theme-doc-footer-edit-meta-row',
		'#docusaurus_skipToContent_fallback + nav'
	];
	selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));

	// 2. 移除宽度限制，让内容自适应 PDF
	const mainWrapper = document.querySelector('.main-wrapper');
	if(mainWrapper) mainWrapper.style.maxWidth = 'none';
	
	const docItemContainer = document.querySelector('.theme-doc-item-container');
	if(docItemContainer) {
		docItemContainer.style.maxWidth = 'none';
		docItemContainer.style.padding = '0';
	}

	// 3. 强制展开所有 details 标签
	document.querySelectorAll('details').forEach(e => e.open = true);

	// 4. 调整页边距
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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🔍 正在扫描 momi.gtemc.cn Wiki 目录...")
	urls := scanLinks(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 发现 %d 个有效页面，开始生成 PDF...\n", len(uniqueUrls))

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
	fmt.Printf("\n✨ 任务完成！\n⏱️ 总耗时: %s\n📄 输出文件: %s\n", time.Since(start), FinalPDF)
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
			network.SetBlockedURLs([]string{"*.woff*", "*.ttf", "*google-analytics*", "*analytics.js*"}),
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("article"), // 等待文章主体加载
			chromedp.Sleep(2*time.Second),  // 给图片留出加载时间
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(true). // 开启背景以保留代码块底色
					WithPaperWidth(8.27).      // A4
					WithPaperHeight(11.69).
					Do(ctx)
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
		fmt.Printf("📄 [%d/%d] 已完成: %s\n", t.ID+1, MaxConcurrent, t.URL)
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
		
		// 格式化 URL，移除结尾斜杠
		cleanCurr := strings.TrimSuffix(curr, "/")
		if visited[cleanCurr] { continue }
		visited[cleanCurr] = true
		
		// 只有包含 /docs/ 的页面通常才是内容页
		if strings.Contains(cleanCurr, "/docs/") || cleanCurr == BaseURL {
			links = append(links, cleanCurr)
		}

		var res []string
		tCtx, tCancel := context.WithTimeout(ctx, 15*time.Second)
		chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.WaitReady("main"),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &res),
		)
		tCancel()

		for _, l := range res {
			u, err := url.Parse(l)
			if err != nil { continue }
			u.Fragment = "" // 移除锚点
			u.RawQuery = "" // 移除参数
			full := strings.TrimSuffix(u.String(), "/")
			
			// 只爬取同站链接，且排除掉 category 这种目录索引页
			if strings.HasPrefix(full, BaseURL) && 
			   !visited[full] && 
			   !strings.Contains(full, "/category/") {
				toVisit = append(toVisit, full)
			}
		}
	}
	return links
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	fmt.Printf("📚 正在合并 %d 个 PDF 页面...\n", len(results))
	
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}

	conf := model.NewDefaultConfiguration()
	// 使用 Relaxed 模式，因为 Docusaurus 产生的 PDF 结构可能较复杂
	conf.ValidationMode = model.ValidationRelaxed

	if err := api.MergeCreateFile(inFiles, FinalPDF, false, conf); err != nil {
		log.Fatalf("合并 PDF 出错: %v", err)
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
