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
	MaxConcurrent = 4
)

// CleanScript removes navigation, sidebar, footer, etc.
const CleanScript = `
	document.querySelectorAll('nav, .sidebar, .navbar, footer, script, iframe, .theme-container > .navbar').forEach(e => e.remove());
	document.querySelectorAll('details').forEach(e => e.open = true);
	document.body.style.padding = '0px';
	document.body.style.margin = '20px';
	document.body.style.backgroundColor = 'white';
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

	// Configure Chrome
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		// Key fix: Large window size to ensure sidebar links are rendered
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	fmt.Println("🕷️ Starting BFS Crawler (Recursive Scan)...")
	
	// Step 1: Recursively crawl all links
	urls := crawlAllPages(allocCtx)
	
	// Deduplicate and sort
	uniqueUrls := uniqueAndSort(urls)
	fmt.Printf("✅ Final Count: %d unique pages found.\n", len(uniqueUrls))

	if len(uniqueUrls) == 0 {
		log.Fatal("❌ No pages found. Check network or BaseURL.")
	}

	// Step 2: Concurrent Rendering
	taskChan := make(chan Task, len(uniqueUrls))
	resChan := make(chan Result, len(uniqueUrls))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < MaxConcurrent; i++ {
		wg.Add(1)
		go worker(allocCtx, taskChan, resChan, &wg)
	}

	// Send tasks
	for i, u := range uniqueUrls {
		taskChan <- Task{ID: i, URL: u}
	}
	close(taskChan)

	// Wait for completion
	wg.Wait()
	close(resChan)

	// Step 3: Collect results
	var results []Result
	for r := range resChan {
		results = append(results, r)
	}
	// Sort by ID to maintain scan order
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })

	// Step 4: Merge
	mergePDFs(results)

	fmt.Printf("🏆 Done! Duration: %s | File: %s\n", time.Since(start), FinalPDF)
	os.RemoveAll(OutDir)
}

// crawlAllPages implements a Breadth-First Search to find all links
func crawlAllPages(rootCtx context.Context) []string {
	// Use a separate context for crawling to avoid interference
	ctx, cancel := chromedp.NewContext(rootCtx)
	defer cancel()

	// Queue for BFS
	queue := []string{BaseURL}
	
	// Track visited URLs to prevent cycles
	visited := make(map[string]bool)
	visited[BaseURL] = true
	
	// List of all discovered URLs in order
	var allLinks []string

	for len(queue) > 0 {
		// Pop the first URL
		currentURL := queue[0]
		queue = queue[1:]
		
		allLinks = append(allLinks, currentURL)
		fmt.Printf("🔍 Scanning [%d found]: %s\n", len(allLinks), currentURL)

		// Extract links from the current page
		foundLinks := extractLinks(ctx, currentURL)
		
		for _, link := range foundLinks {
			// Normalize link: remove fragment, trailing slash
			u, err := url.Parse(link)
			if err != nil { continue }
			u.Fragment = ""
			normalized := strings.TrimSuffix(u.String(), "/")

			// Check if it's internal and not visited
			if strings.HasPrefix(normalized, BaseURL) && !visited[normalized] {
				visited[normalized] = true
				queue = append(queue, normalized)
			}
		}
	}
	return allLinks
}

func extractLinks(ctx context.Context, targetURL string) []string {
	// Set a timeout for each page scan
	tCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var res []string
	err := chromedp.Run(tCtx,
		chromedp.Navigate(targetURL),
		// Wait for body to ensure DOM is loaded
		chromedp.WaitReady("body"),
		// Small sleep to allow JS frameworks (VuePress/React) to render sidebar
		chromedp.Sleep(500*time.Millisecond),
		// Extract all hrefs
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a[href]')).map(a => a.href)`, &res),
	)
	
	if err != nil {
		fmt.Printf("⚠️ Failed to scan %s: %v\n", targetURL, err)
		return []string{}
	}
	return res
}

func worker(parentCtx context.Context, tasks <-chan Task, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := chromedp.NewContext(parentCtx)
	defer cancel()

	// Block fonts and media to save bandwidth, but ALLOW images/css
	chromedp.Run(ctx, network.Enable(), network.SetBlockedURLs([]string{
		"*.woff", "*.woff2", "*.ttf", "*.otf", "*.mp4", "*google-analytics*",
	}))

	for t := range tasks {
		var buf []byte
		tCtx, tCancel := context.WithTimeout(ctx, 45*time.Second)
		
		err := chromedp.Run(tCtx,
			chromedp.Navigate(t.URL),
			chromedp.WaitReady("body"),
			chromedp.Sleep(1*time.Second), // Wait for lazy-loaded images
			chromedp.Evaluate(CleanScript, nil),
			chromedp.ActionFunc(func(ctx context.Context) error {
				var err error
				// Assign to outer 'buf'
				buf, _, err = page.PrintToPDF().
					WithPrintBackground(false). // Keep content images, remove background colors
					WithPaperWidth(8.27).WithPaperHeight(11.69).
					WithMarginTop(0.3).WithMarginBottom(0.3).
					WithMarginLeft(0.3).WithMarginRight(0.3).
					Do(ctx)
				return err
			}),
		)
		tCancel()

		if err != nil {
			fmt.Printf("⚠️ Render Failed: %s\n", t.URL)
			continue
		}

		path := filepath.Join(OutDir, fmt.Sprintf("%03d.pdf", t.ID))
		os.WriteFile(path, buf, 0644)
		results <- Result{ID: t.ID, Path: path}
		fmt.Printf("📄 [%d] Saved: %s\n", t.ID, t.URL)
	}
}

func mergePDFs(results []Result) {
	if len(results) == 0 { return }
	fmt.Println("📚 Merging PDFs...")
	var inFiles []string
	for _, r := range results {
		inFiles = append(inFiles, r.Path)
	}
	// Pass nil config to use defaults
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
