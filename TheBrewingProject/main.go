package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

// 清理无效内部锚点链接，解决 Pandoc 转换 PDF 时的报错
func cleanInternalLinks(content string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(#[^\)]+\)`)
	return re.ReplaceAllString(content, "$1")
}

func main() {
	outputFile := "TheBrewingProject_Wiki.md"
	f, _ := os.Create(outputFile)
	defer f.Close()

	// 1. 写入文档元数据
	f.WriteString("---\n")
	f.WriteString("title: The Brewing Project 完整 Wiki 手册\n")
	f.WriteString("author: 艾尔岚 (Ellan) 自动化助手\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	// 定义 Hangar 的特定参数
	baseURL := "https://hangar.papermc.io"
	startURL := "https://hangar.papermc.io/BreweryTeam/TheBrewingProject/pages/Wiki"
	projectPath := "/BreweryTeam/TheBrewingProject/pages/"

	visited := make(map[string]bool)
	c := colly.NewCollector(
		colly.AllowedDomains("hangar.papermc.io"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	// 2. 提取正文内容
	// Hangar 的文档主要位于 .markdown-content 或 .project-page 内
	c.OnHTML(".project-page, .markdown-content", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] {
			return
		}
		visited[url] = true

		// 提取标题：优先取正文 H1，若无则取 URL 最后一段
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
			title = parts[len(parts)-1]
		}
		
		fmt.Printf("正在抓取页面: %s\n", title)

		// 修复图片路径，将相对路径转换为绝对 URL
		e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
			imgSrc, exists := s.Attr("src")
			if exists && strings.HasPrefix(imgSrc, "/") {
				s.SetAttr("src", baseURL+imgSrc)
			}
		})

		// 标注上下文（对 EcoBridge 处理酿酒逻辑非常有用）
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		html, _ := e.DOM.Html()
		markdown, _ := converter.ConvertString(html)
		
		// 清理导致 PDF 报错的无效内部锚点
		finalMarkdown := cleanInternalLinks(markdown)

		f.WriteString(fmt.Sprintf("# %s\n\n", title))
		f.WriteString(fmt.Sprintf("> 来源: %s\n\n", url))
		f.WriteString(finalMarkdown)
		f.WriteString("\n\n\\newpage\n\n")
	})

	// 3. 递归寻找 Wiki 页面链接（侧边栏或页面内）
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// 确保只爬取该项目的 pages 目录下的链接
		if strings.HasPrefix(link, projectPath) && !strings.Contains(link, "#") {
			fullLink := e.Request.AbsoluteURL(link)
			if !visited[fullLink] {
				e.Request.Visit(fullLink)
			}
		}
	})

	fmt.Println("🚀 启动 Hangar 专用爬虫...")
	c.Visit(startURL)
	c.Wait()
	fmt.Println("✨ 抓取完成！文件已保存为:", outputFile)
}