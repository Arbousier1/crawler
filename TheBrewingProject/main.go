package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

func main() {
	outputFile := "TheBrewingProject_Wiki.md"
	f, _ := os.Create(outputFile)
	defer f.Close()

	// 写入 PDF 元数据
	f.WriteString("---\ntitle: The Brewing Project 官方百科\nauthor: 艾尔岚开发组\ntoc: true\nlang: zh-CN\n---\n\n")

	visited := make(map[string]bool)
	// 创建爬虫实例
	c := colly.NewCollector(
		colly.AllowedDomains("hangar.papermc.io"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	// 1. 提取正文逻辑
	// Hangar 的文档正文通常在 .markdown-content 或 .project-page 内
	c.OnHTML(".markdown-content, .project-page, .markdown-body", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] {
			return
		}
		visited[url] = true

		// 获取标题：优先找 H1，找不到则用 URL 最后一段
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
			title = parts[len(parts)-1]
		}

		fmt.Printf("✅ 正在提取章节: %s\n", title)

		// 修复相对路径图片
		e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
			src, _ := s.Attr("src")
			if strings.HasPrefix(src, "/") {
				s.SetAttr("src", "https://hangar.papermc.io"+src)
			}
		})

		html, _ := e.DOM.Html()
		markdown, _ := converter.ConvertString(html)

		f.WriteString(fmt.Sprintf("# %s\n\n%s\n\n\\newpage\n\n", title, markdown))
	})

	// 2. 发现侧边栏链接逻辑
	// 匹配侧边栏或页面中所有指向 /pages/ 的内部链接
	c.OnHTML("a[href*='/pages/']", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		// 确保链接属于该插件的文档范围，且排除锚点
		if strings.Contains(link, "/BreweryTeam/TheBrewingProject/pages/") && !strings.Contains(link, "#") {
			c.Visit(link)
		}
	})

	fmt.Println("🚀 正在从 Hangar 网站深度爬取 BrewingProject Wiki...")
	c.Visit("https://hangar.papermc.io/BreweryTeam/TheBrewingProject/pages/Wiki")
	c.Wait()
	fmt.Println("✨ 抓取完成，文件已保存为:", outputFile)
}

