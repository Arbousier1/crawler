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

const CleanScript = `
	// 移除导航、侧边栏、右侧目录、翻页、页脚
	const selectors = ['.navbar', '.theme-doc-sidebar-container', '.table-of-contents', '.pagination-nav', 'footer'];
	selectors.forEach(s => document.querySelectorAll(s).forEach(e => e.remove()));

	// 移除最大宽度限制，让 PDF 铺满 A4
	const mainWrapper = document.querySelector('.main-wrapper');
	if(mainWrapper) mainWrapper.style.maxWidth = 'none';
	const docContainer = document.querySelector('.theme-doc-item-container');
	if(docContainer) docContainer.style.maxWidth = 'none';

	// 强制展开所有细节
	document.querySelectorAll('details').forEach(e => e.open = true);
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

	fmt.Println("🔍 扫描 MOMI Wiki 链接...")
	urls := scanLinks(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 发现 %d 个页面，开始并发生成...\n", len(uniqueUrls))

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
	fmt.Printf("\n✨ 任务完成！总耗时: %s\n", time.Since(start))
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
			network.SetBlockedURLs([]string{"*.woff*", "*.ttf", "*google-analytics*"}),
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("article"),
			chromedp.Sleep(2*time.Second), 
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(true). // 保留代码块底色
					WithPaperWidth(8.27).      // A4 Width
					WithPaperHeight(11.69).    // A4 Height
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
		fmt.Printf("📄 [%d] 已处理: %s\n", t.ID, t.URL)
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
		cleanCurr := strings.TrimSuffix(curr, "/")
		if visited[cleanCurr] { continue }
		visited[cleanCurr] = true
		
		if strings.Contains(cleanCurr, "/docs/") || cleanCurr == BaseURL {
			links = append(links, cleanCurr)
		}

		var res []string
		tCtx, tCancel := context.WithTimeout(ctx, 15*time.Second)
		chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &res),
		)
		tCancel()

		for _, l := range res {
			u, err := url.Parse(l)
			if err != nil || !strings.HasPrefix(l, BaseURL) { continue }
			u.Fragment = ""
			u.RawQuery = ""
			full := strings.TrimSuffix(u.String(), "/")
			if !visited[full] && !strings.Contains(full, "/category/") {
				toVisit = append(toVisit, full)
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
	conf.ValidationMode = model.ValidationRelaxed // 核心修复：旧版 ValidationNone 已移除
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
