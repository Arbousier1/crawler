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
	BaseURL       = "https://xiao-momi.github.io/craft-engine-wiki/"
	OutDir        = "dist"
	FinalPDF      = "Wiki_Multimodal_AI.pdf"
	// 并发数：因为要下载图片，内存压力变大，GitHub Action 建议保守点设为 3-4
	MaxConcurrent = 4 
)

// DOM 净化脚本：保留图片，但删除导航和无用元素
const CleanScript = `
	// 移除导航、侧边栏、页脚、脚本、iframe (保留样式以维持布局)
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
	os.RemoveAll(OutDir)
	os.MkdirAll(OutDir, 0755)

	// 1. 启动浏览器
	// 关键：不能禁用图片引擎了 (去掉了 imagesEnabled=false)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true), // CI 环境依然禁用 GPU 以稳为主
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
	// 去重并排序，保证基础顺序
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

	// 4. 收集结果
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	// 5. 极速合并
	mergePDFs(results)

	fmt.Printf("🏆 多模态工程完成！耗时: %s | 文件: %s\n", time.Since(start), FinalPDF)
	// 清理临时文件
	os.RemoveAll(OutDir)
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	// 【关键调整】网络拦截策略
	// 不再拦截 CSS 和图片，只拦截字体、媒体和统计脚本
	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.otf", // 字体文件贼大，AI不需要
		"*.mp4", "*.webm", "*.mp3",            // 媒体文件
		"*google-analytics*", "*hm.baidu*",    // 统计脚本
	}))

	for t := range tasks {
		var buf []byte
		// 因为要加载图片，超时时间稍微给多点
		tCtx, tCancel := context.WithTimeout(ctx, 60*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			// 【关键】必须等待网络空闲 (networkIdle)，确保图片加载完毕
			chromedp.WaitReady("body"),
			chromedp.Sleep(1*time.Second), // 额外缓冲，确保懒加载图片出现

			// 执行 DOM 手术
			chromedp.Evaluate(CleanScript, nil),

			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				buf, _, err = page.PrintToPDF().
					// 【核心关键】false = 不打印背景色/背景图，但保留正文图片
					WithPrintBackground(false). 
					WithPaperWidth(8.27).
					WithPaperHeight(11.69).
					// 边距设置小一点，让内容更紧凑
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
		fmt.Printf("🖼️ [%d/%d] 已渲染(含图): %s\n", t.ID+1, cap(tasks), t.URL)
	}
}

func scanLinks(ctx context.Context) []string {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	
	var links []string
	toVisit := []string{BaseURL}
	visited := make(map[string]bool)
	
	// 简单的 BFS 扫描
	for len(toVisit) > 0 {
		curr := toVisit[0]
		toVisit = toVisit[1:]
		if visited[curr] { continue }
		visited[curr] = true
		links = append(links, curr)

		var res []string
		// 扫描时不需要加载图片，可以快点
		tCtx, tCancel := context.WithTimeout(ctx, 15*time.Second)
		chromedp.Run(tCtx, 
			chromedp.Navigate(curr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a=>a.href)`, &res),
		)
		tCancel()

		for _, l := range res {
			u, err := url.Parse(l)
			if err != nil { continue }
			u.Fragment = "" // 去掉锚点
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
	fmt.Println("📚 正在进行内存级 PDF 合并...")
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationNone
	if err := api.MergeCreateFile(inFiles, FinalPDF, false, conf); err != nil {
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
