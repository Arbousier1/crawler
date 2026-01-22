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

// Config 配置项
const (
	// 目标入口 URL
	TargetURL = "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops"
	// 保存 Markdown 的目录
	OutputDir = "./knowledge_base"
)

func main() {
	// 1. 初始化输出目录
	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		log.Fatalf("无法创建输出目录: %v", err)
	}

	// 2. 配置 Chrome 启动选项
	// 注意：在 GitHub Actions (Docker环境) 中，no-sandbox 是必须的
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// 伪装 User-Agent，防止简单的反爬拦截
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		
		// 启用无头模式 (不显示 UI)
		chromedp.Flag("headless", true),
		
		// 禁用 GPU 加速 (服务器环境通常不需要)
		chromedp.DisableGPU,
		
		// ⚠️ CI/CD 环境关键配置 ⚠️
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true), // 防止在 Docker 中内存不足崩溃
	)

	// 创建分配器上下文
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建浏览器上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置总超时时间 (例如 15 分钟)，防止脚本卡死
	ctx, cancel = context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	log.Println("🚀 开始初始化浏览器并扫描目录...")

	// 3. 获取侧边栏所有链接
	links, err := fetchSidebarLinks(ctx, TargetURL)
	if err != nil {
		log.Fatalf("获取目录结构失败: %v", err)
	}

	log.Printf("🔍 发现 %d 个页面，准备开始爬取...\n", len(links))

	// 4. 初始化 Markdown 转换器
	converter := md.NewConverter("", true, nil)

	// 5. 遍历链接并爬取
	for i, link := range links {
		// 简单的速率限制，防止请求过快被封
		time.Sleep(2 * time.Second)

		log.Printf("[%d/%d] 正在处理: %s", i+1, len(links), link)

		contentHTML, title, err := fetchPageContent(ctx, link)
		if err != nil {
			log.Printf("❌ 获取页面失败 [%s]: %v", link, err)
			continue
		}

		// HTML 转 Markdown
		markdown, err := converter.ConvertString(contentHTML)
		if err != nil {
			log.Printf("⚠️ 转换 Markdown 失败 [%s]: %v", title, err)
			continue
		}

		// 拼接最终文件内容 (包含元数据头，利于 AI 溯源)
		finalContent := fmt.Sprintf("# %s\n\n> Source: %s\n\n---\n\n%s", title, link, markdown)

		// 保存文件
		filename := cleanFilename(title) + ".md"
		savePath := filepath.Join(OutputDir, filename)

		if err := os.WriteFile(savePath, []byte(finalContent), 0644); err != nil {
			log.Printf("💾 保存文件失败: %v", err)
		} else {
			log.Printf("✅ 已保存: %s", filename)
		}
	}

	log.Println("🎉 全部完成！文件已保存在:", OutputDir)
}

// fetchSidebarLinks 访问主页并解析侧边栏链接
func fetchSidebarLinks(ctx context.Context, urlStr string) ([]string, error) {
	var htmlContent string
	
	// 设置较长的超时以等待页面初次加载
	scanCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(scanCtx,
		network.Enable(),
		chromedp.Navigate(urlStr),
		// 等待 GitBook 的侧边栏或主内容加载完成
		chromedp.WaitVisible("body", chromedp.ByQuery), 
		// 稍微多等一下确保 JS 执行完毕
		chromedp.Sleep(2*time.Second), 
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

	// GitBook 侧边栏通常在 <nav> 标签下
	doc.Find("nav a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			// 解析绝对路径
			parsedLink, err := baseURL.Parse(href)
			if err != nil {
				return
			}
			fullLink := parsedLink.String()

			// 过滤逻辑：只抓取同域名的链接，且去重
			if parsedLink.Host == baseURL.Host && !seen[fullLink] {
				// 排除一些显然不是文档的链接 (可选)
				if !strings.Contains(fullLink, "/edit/") {
					links = append(links, fullLink)
					seen[fullLink] = true
				}
			}
		}
	})

	// 如果 nav 没抓到，尝试兜底抓取当前页面所有同域链接 (GitBook 结构多变)
	if len(links) == 0 {
		log.Println("⚠️ 未在 nav 中发现链接，尝试扫描全文链接...")
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.HasPrefix(href, "/") {
				parsedLink, _ := baseURL.Parse(href)
				fullLink := parsedLink.String()
				if !seen[fullLink] {
					links = append(links, fullLink)
					seen[fullLink] = true
				}
			}
		})
	}

	return links, nil
}

// fetchPageContent 获取单页面的正文 HTML 和 标题
func fetchPageContent(ctx context.Context, urlStr string) (string, string, error) {
	var htmlContent string
	
	// 单页超时控制
	pageCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(pageCtx,
		chromedp.Navigate(urlStr),
		// 等待 main 标签，这是 GitBook 正文通常所在的位置
		chromedp.WaitVisible("main", chromedp.ByQuery),
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
		// 如果没有 h1，尝试从 title 标签拿
		title = doc.Find("title").Text()
		// 清理类似 "Page Title - GitBook" 的后缀
		if idx := strings.Index(title, " - "); idx != -1 {
			title = title[:idx]
		}
	}
	if title == "" {
		title = "Untitled_" + fmt.Sprintf("%d", time.Now().Unix())
	}

	// 提取正文区域
	mainSelection := doc.Find("main")
	
	// 清理无用元素，减少 AI 干扰
	mainSelection.Find("script, style, iframe, noscript, svg, button").Remove()
	// 移除 GitBook 底部翻页导航
	mainSelection.Find("a[class*='pagination']").Remove() 

	contentHTML, err := mainSelection.Html()
	if err != nil {
		return "", title, err
	}

	return contentHTML, title, nil
}

// cleanFilename 处理非法文件名字符
func cleanFilename(name string) string {
	// 替换常见非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "-")
	}
	// 移除首尾空格和过多的横杠
	result = strings.TrimSpace(result)
	return result
}
