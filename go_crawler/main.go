package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const BaseURL = "https://xiao-momi.github.io/craft-engine-wiki/"
const OutDir = "temp_pdfs"

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
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 1. 扫描链接 (简单演示，实际可用递归)
	urls := []string{BaseURL} // 实际生产中建议加入递归扫描逻辑
	
	// 2. 并发抓取
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []PageMeta
	sem := make(chan struct{}, 5) // 限制并发

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, targetURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ctx, _ := chromedp.NewContext(allocCtx)
			var title string
			var buf []byte
			
			chromedp.Run(ctx,
				chromedp.Navigate(targetURL),
				chromedp.WaitReady("body"),
				chromedp.Title(&title),
				chromedp.ActionFunc(func(ctx context.Context) error {
					var err error
					buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
					return err
				}),
			)

			fileName := fmt.Sprintf("page_%03d.pdf", idx)
			filePath := filepath.Join(OutDir, fileName)
			os.WriteFile(filePath, buf, 0644)

			mu.Lock()
			results = append(results, PageMeta{ID: idx, Title: title, URL: targetURL, Path: filePath})
			mu.Unlock()
			fmt.Printf("Done: %s\n", targetURL)
		}(i, u)
	}
	wg.Wait()

	// 保存元数据供 Python 使用
	metaJson, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile("metadata.json", metaJson, 0644)
}
