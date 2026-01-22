package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

func main() {
	// 最终合并的 Markdown 文件名
	combinedFile := "customcrops_wiki.md"
	f, err := os.Create(combinedFile)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 写入 Pandoc 识别的元数据头部 (用于生成 PDF 封面和标题)
	f.WriteString("---\n")
	f.WriteString("title: CustomCrops 完整插件文档\n")
	f.WriteString("author: 自动化知识库助手\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("toc-title: 目录\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	visited := make(map[string]bool)
	c := colly.NewCollector(
		colly.AllowedDomains("mo-mi.gitbook.io", "gitbook.io"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	c.OnHTML("main", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] || !strings.Contains(url, "customcrops") {
			return
		}
		visited[url] = true

		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			title = "未命名章节"
		}
		fmt.Printf("正在提取章节: %s\n", title)

		// 标注代码块所属页面
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		htmlContent, _ := e.DOM.Html()
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			return
		}

		// 写入 Markdown 内容
		// # 是一级标题，Pandoc 会据此生成目录
		f.WriteString(fmt.Sprintf("# %s\n\n", title))
		f.WriteString(fmt.Sprintf("> 原始链接: [%s](%s)\n\n", url, url))
		f.WriteString(markdown)
		f.WriteString("\n\n\\newpage\n\n") // 强制 PDF 换页
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Contains(link, "mo-mi.gitbook.io/xiaomomi-plugins/customcrops") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	fmt.Println("🚀 启动合并抓取程序...")
	c.Visit("https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops")
	c.Wait()
	fmt.Println("✨ Markdown 构建完成，准备转换为 PDF...")
}
