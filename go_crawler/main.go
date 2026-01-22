package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Config
const (
	TargetURL = "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops"
	OutputDir = "./knowledge_base"
)

func main() {
	// 1. 准备输出目录
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		log.Fatalf("❌ 无法创建目录: %v", err)
	}

	// 2. 配置 Chrome (针对 GitHub Actions 优化)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
		chromedp.DisableGPU,
		// ⚠️ CI 环境关键配置：防止 crashing
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// 设置全局超时 (15分钟)
	ctx, cancel = context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	log.Println("🚀 启动爬虫，正在分析目录...")

	// 3. 获取链接
	links, err := fetchSidebarLinks(ctx, TargetURL)
	if err != nil {
		log.Fatalf("❌ 获取目录失败: %v", err)
	}

	log.Printf("🔍 发现 %d 个页面，开始爬取内容...\n", len(links))

	// 4. 遍历爬取
	converter := md.NewConverter("", true, nil)

	for i, link := range links {
		// 避免请求过快
		time.Sleep(2 * time.Second)
		log.Printf("[%d/%d] 处理: %s", i+1, len(links), link)

		html, title, err := fetchPageContent(ctx, link)
		if err != nil {
			log.Printf("⚠️ 跳过页面 [%s]: %v", link, err)
			continue
		}

		// 转 Markdown
		markdown, err := converter.ConvertString(html)
		if err != nil {
			log.Printf("⚠️ 转换失败 [%s]: %v", title, err)
			continue
		}

		// 拼接内容
		fileContent := fmt.Sprintf("# %s\n\n> Original URL: %s\n\n---\n\n%s", title, link, markdown)
		
		// 保存
		filename := cleanFilename(title) + ".md"
		if err := os.WriteFile(filepath.Join(OutputDir, filename), []byte(fileContent), 0644); err != nil {
			log.Printf("❌ 保存失败: %v", err)
		}
	}

	log.Println("✅ 任务完成！所有文件已保存至:", OutputDir)
}

func fetchSidebarLinks(ctx context.Context, urlStr string) ([]string, error) {
	var htmlContent string
	// 给予足够的时间加载侧边栏
	tCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	err := chromedp.Run(tCtx,
		network.Enable(),
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // 等待 JS 渲染
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var links []string
	seen := make(map[string]bool)
	baseURL, _ := url.Parse(urlStr)

	// 抓取逻辑：优先查找 nav 标签
	doc.Find("nav a").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			resolveLink(&links, seen, baseURL, href)
		}
	})

	// 兜底逻辑：如果 nav 没抓到，抓取所有同域链接
	if len(links) == 0 {
		doc.Find("a").Each(func(_ int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				resolveLink(&links, seen, baseURL, href)
			}
		})
	}

	return links, nil
}

func resolveLink(links *[]string, seen map[string]bool, base *url.URL, href string) {
	// 解析绝对路径
	u, err := base.Parse(href)
	if err != nil {
		return
	}
	// 只保留同域名下的内容
	if u.Host == base.Host && !seen[u.String()] {
		// 排除非文档链接
		if !strings.Contains(u.String(), "/edit/") && !strings.Contains(u.String(), "/history/") {
			*links = append(*links, u.String())
			seen[u.String()] = true
		}
	}
}

func fetchPageContent(ctx context.Context, urlStr string) (string, string, error) {
	var htmlContent string
	tCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(tCtx,
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("main", chromedp.ByQuery), // 只要主内容出来即可
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", err
	}

	// 提取标题
	title := strings.TrimSpace(doc.Find("h1").First().Text())
	if title == "" {
		title = doc.Find("title").Text()
	}
	if title == "" {
		title = "Untitled"
	}

	// 提取正文
	main := doc.Find("main")
	// 清洗干扰元素
	main.Find("script, style, noscript, iframe, svg, button").Remove()
	main.Find("a[class*='pagination']").Remove() // 移除底部翻页按钮

	content, err := main.Html()
	return content, title, err
}

func cleanFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\t"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "-")
	}
	return strings.TrimSpace(name)
}
