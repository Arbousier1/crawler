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

// 清理无效锚点，防止 PDF 报错
func cleanInternalLinks(content string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(#[^\)]+\)`)
	return re.ReplaceAllString(content, "$1")
}

func main() {
	outputFile := "TheBrewingProject_Wiki.md"
	f, _ := os.Create(outputFile)
	defer f.Close()

	// 写入元数据
	f.WriteString("---\n")
	f.WriteString("title: The Brewing Project 官方 Wiki 百科\n")
	f.WriteString("author: 艾尔岚 (Ellan) 开发组\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	baseURL := "https://hangar.papermc.io"
	// Hangar 页面内容通常在这个路径前缀下
	projectPath := "/BreweryTeam/TheBrewingProject/pages"

	visited := make(map[string]bool)
	c := colly.NewCollector(
		colly.AllowedDomains("hangar.papermc.io"),
		// 模拟真实浏览器，防止被 Hangar 的防火墙拦截返回空页面
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	// 核心逻辑：提取 Hangar 的 Wiki 内容
	// Hangar 的文档主要包裹在 .project-page 或 .markdown-content 中
	c.OnHTML(".project-page, .markdown-content, .markdown-body", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] { return }
		visited[url] = true

		// 提取标题
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
			title = parts[len(parts)-1]
		}
		
		fmt.Printf("成功抓取页面: %s\n", title)

		// 修复相对图片路径
		e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
			imgSrc, _ := s.Attr("src")
			if strings.HasPrefix(imgSrc, "/") {
				s.SetAttr("src", baseURL+imgSrc)
			}
		})

		html, _ := e.DOM.Html()
		markdown, _ := converter.ConvertString(html)
		finalMarkdown := cleanInternalLinks(markdown)

		f.WriteString(fmt.Sprintf("# %s\n\n", title))
		f.WriteString(fmt.Sprintf("> 来源: %s\n\n", url))
		f.WriteString(finalMarkdown)
		f.WriteString("\n\n\\newpage\n\n")
	})

	// 关键逻辑：寻找导航栏中的所有子页面链接
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// 只要链接包含项目路径，就尝试去访问
		if strings.Contains(link, projectPath) && !strings.Contains(link, "#") {
			fullLink := e.Request.AbsoluteURL(link)
			if !visited[fullLink] {
				e.Request.Visit(fullLink)
			}
		}
	})

	fmt.Println("🚀 正在深度爬取 Hangar Wiki...")
	c.Visit("https://hangar.papermc.io/BreweryTeam/TheBrewingProject/pages/Wiki")
	c.Wait()
}
