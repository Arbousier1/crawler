package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const BaseURL = "https://xiao-momi.github.io/craft-engine-wiki/"
const OutDir = "../temp_pdfs"

type PageMeta struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Path  string `json:"path"`
}

func main() {
	os.MkdirAll(OutDir, 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 1. 扫描所有链接
	urls := scanAllLinks(allocCtx)
	fmt.Printf("✅ 扫描完成：共发现 %d 个页面\n", len(urls))

	// 2. 并发生成 PDF
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]PageMeta, 0)
	sem := make(chan struct{}, 5) 

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, targetURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ctx, _ := chromedp.NewContext(allocCtx)
			var title string
			var buf []byte
			
			err := chromedp.Run(ctx,
				chromedp.Navigate(targetURL),
				chromedp.WaitReady("body"),
				chromedp.Sleep(2*time.Second), 
				chromedp.Title(&title),
				chromedp.ActionFunc(func(ctx context.Context) error {
					var err error
					buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
					return err
				}),
			)
			if err != nil {
				return
			}

			fileName := fName(idx)
			filePath := filepath.Join(OutDir, fileName)
			os.WriteFile(filePath, buf, 0644)

			mu.Lock()
			results = append(results, PageMeta{ID: idx, Title: title, URL: targetURL, Path: fileName})
			mu.Unlock()
			fmt.Printf("🚀 [%d/%d] 已转换: %s\n", idx+1, len(urls), targetURL)
		}(i, u)
	}
	wg.Wait()

	metaData, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile("../metadata.json", metaData, 0644)
}

func fName(i int) string {
	return fmt.Sprintf("page_%03d.pdf", i)
}

func scanAllLinks(allocCtx context.Context) []string {
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	visited := make(map[string]bool)
	toVisit := []string{BaseURL}
	var all []string

	for len(toVisit) > 0 {
		curr := toVisit[0]
		toVisit = toVisit[1:]
		if visited[curr] { continue }
		visited[curr] = true
		all = append(all, curr)

		var res []string
		chromedp.Run(ctx,
			chromedp.Navigate(curr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a => a.href)`, &res),
		)
		for _, link := range res {
			u, _ := url.Parse(link)
			u.Fragment = ""
			full := strings.TrimSuffix(u.String(), "/")
			if strings.HasPrefix(full, BaseURL) && !visited[full] {
				toVisit = append(toVisit, full)
			}
		}
	}
	return all
}
