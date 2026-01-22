package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Config 配置项
const (
	TargetURL    = "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops" // 目标入口
	OutputDir    = "./knowledge_base"                                      // 保存目录
	WaitSelector = "main"                                                  // GitBook 内容通常在 main 标签中
)

func main() {
	// 1. 初始化输出目录
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		log.Fatal(err)
	}

	// 2. 配置 Chrome (Headless 模式)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true), // 如果想看浏览器运行，改为 false
		chromedp.DisableGPU,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时时间，防止脚本无限挂起
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	log.Println("🚀 开始扫描目录结构...")

	// 3. 获取所有侧边栏链接
	links, err := fetchSidebarLinks(ctx, TargetURL)
	if err != nil {
		log.Fatalf("获取目录失败: %v", err)
	}

	log.Printf("发现 %d 个页面，开始爬取内容...\n", len(links))

	// 4. 遍历链接并爬取内容
	converter := md.NewConverter("", true, nil)

	for i, link := range links {
		// 简单的防封禁策略：休眠 1-3 秒
		time.Sleep(2 * time.Second)

		log.Printf("[%d/%d] 处理: %s", i+1, len(links), link)
		
		content, title, err := fetchPageContent(ctx, link)
		if err != nil {
			log.Printf("❌ 失败 %s: %v", link, err)
			continue
		}

		// 5. 转换为 Markdown
		markdown, err := converter.ConvertString(content)
		if err != nil {
			log.Printf("⚠️ 转换 Markdown 失败: %v", err)
			continue
		}

		// 添加原文链接到头部，方便追溯
		finalMD := fmt.Sprintf("# %s\n\nSource: %s\n\n%s", title, link, markdown)

		// 6. 保存文件
		filename := cleanFilename(title) + ".md"
		savePath := filepath.Join(OutputDir, filename)
		if err := os.WriteFile(savePath, []byte(finalMD), 0644); err != nil {
			log.Printf("无法保存文件: %v", err)
		} else {
			log.Printf("✅ 已保存: %s", filename)
		}
	}
	
	log.Println("🎉 爬取完成！所有文件已保存至", OutputDir)
}

// fetchSidebarLinks 获取侧边栏的所有链接
func fetchSidebarLinks(ctx context.Context, urlStr string) ([]string, error) {
	var htmlContent string
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(urlStr),
		// 等待侧边栏加载，GitBook 的侧边栏通常在 nav 标签或者特定的 div 中
		// 这里等待 main 加载，说明页面大体已经 ok
		chromedp.WaitVisible("main", chromedp.ByQuery),
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

	// GitBook 侧边栏链接通常在 nav 里面
	doc.Find("nav a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			// 处理相对路径
			if strings.HasPrefix(href, "/") {
				// 拼接域名 (这里需要简单的 url parsing，为演示方便硬编码前缀逻辑)
				// 实际 GitBook 往往是 subdomain.gitbook.io
				// 注意：如果 href 是相对当前路径的，这里需要更复杂的 URL Resolve
				// GitBook 通常生成的 href 是相对根目录的，或者是完整的
				if !strings.HasPrefix(href, "http") {
					baseURL := "https://mo-mi.gitbook.io" // 基础域名
					href = baseURL + href
				}
			}
			
			// 只保留本站的链接，排除外部链接
			if strings.Contains(href, "mo-mi.gitbook.io") && !seen[href] {
				links = append(links, href)
				seen[href] = true
			}
		}
	})

	return links, nil
}

// fetchPageContent 获取单个页面的主要内容
func fetchPageContent(ctx context.Context, urlStr string) (string, string, error) {
	var htmlContent string
	// 这里的超时控制单个页面的加载时间
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(urlStr),
		chromedp.WaitVisible("main", chromedp.ByQuery), // 等待正文出现
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", err
	}

	// 获取标题
	title := doc.Find("h1").First().Text()
	if title == "" {
		title = "Untitled"
	}

	// 获取正文 (GitBook 的正文通常在 main 标签里)
	mainContent := doc.Find("main")
	
	// 移除不需要的元素，保持语料干净
	mainContent.Find("script, style, iframe, noscript, nav").Remove()
	
	// 获取 HTML 字符串
	contentHtml, err := mainContent.Html()
	if err != nil {
		return "", title, err
	}

	return contentHtml, title, nil
}

// cleanFilename 清理文件名中的非法字符
func cleanFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\t"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return strings.TrimSpace(name)
}
