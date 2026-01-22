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
	FinalPDF      = "Wiki_Multimodal_AI.pdf"
	// 并发数：根据 GitHub Action 性能建议设为 3-4
	MaxConcurrent = 4 
)

// DOM 净化脚本：保留图片，但删除导航和无用元素
const CleanScript = `
    // 移除导航、侧边栏、页脚、脚本、iframe
    document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe').forEach(e => e.remove());
    // 强制展开详情
    document.querySelectorAll('details').forEach(e => e.open = true);
    // 调整 body 样式以适应 PDF
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
	
	// 初始化目录
	os.RemoveAll(OutDir)
	if err := os.MkdirAll(OutDir, 0755); err != nil {
		log.Fatal(err)
	}

	// 1. 启动浏览器配置
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		// 增加共享内存，防止图片过多导致崩溃
		chromedp.Flag("disable-dev-shm-usage", true), 
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 2. 扫描链接
	fmt.Println("⚡ 正在扫描全站链接...")
	urls := scanLinks(allocCtx)
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ 扫描完成: %d 个唯一页面，开始并发渲染(含图片)...\n", len(uniqueUrls))

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

	// 4. 收集并排序结果
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	// 5. 合并 PDF
	mergePDFs(results)

	fmt.Printf("🏆 任务完成！耗时: %s | 生成文件: %s\n", time.Since(start), FinalPDF)
	
	// 可选：清理临时文件
	// os.RemoveAll(OutDir)
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// 为每个 worker 创建独立的上下文
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	for t := range tasks {
		var buf []byte
		// 渲染单页，超时设为 60s 以保证图片加载
		tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
		
		err := chromedp.Run(tCtx,
			network.Enable(),
			// 拦截非必要资源，节省带宽和内存
			network.SetBlockedURLs([]string{
				"*.woff", "*.woff2", "*.ttf", "*.otf", 
				"*.mp4", "*.webm", "*.mp3",           
				"*google-analytics*", "*hm.baidu*",   
			}),
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(2*time.Second), // 缓冲时间，确保懒加载图片加载完成
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(false).
					WithPaperWidth(8.27).
					WithPaperHeight(11.69).
					WithMarginTop(0.3).WithMarginBottom(0.3).
					WithMarginLeft(0.3).WithMarginRight(0.3).
					Do(ctx)
				return err
			}),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ 渲染失败 [%d]: %s (%v)\n", t.ID, t.URL, err)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		if err := os.WriteFile(path, buf, 0644); err != nil {
			fmt.Printf("⚠️ 保存失败 [%d]: %v\n", t.ID, err)
			continue
		}
		
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("🖼️  [%d] 已渲染: %s\n", t.ID, t.URL)
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
		err := chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &res),
		)
		tCancel()
		
		if err != nil { continue }

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
	if len(results) == 0 {
		fmt.Println("❌ 没有可合并的文件")
		return
	}
	
	fmt.Println("📚 正在进行 PDF 合并...")
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}

	// 关键修复点：使用 NewDefaultConfiguration 并设置 ValidationRelaxed
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	if err := api.MergeCreateFile(inFiles, FinalPDF, false, conf); err != nil {
		log.Fatalf("❌ 合并失败: %v", err)
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
