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
	// 1. 设置合并后的文件名
	combinedFile := "craft_engine_wiki.md"
	f, err := os.Create(combinedFile)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 2. 写入 PDF 元数据头部 (Pandoc 格式)
	f.WriteString("---\n")
	f.WriteString("title: CraftEngine 完整开发指南\n")
	f.WriteString("author: AI 知识库机器人\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("toc-title: 目录\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	visited := make(map[string]bool)
	// 基础路径过滤，确保不爬取外部链接
	basePath := "/craft-engine-wiki/"

	c := colly.NewCollector(
		colly.AllowedDomains("xiao-momi.github.io"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	// 3. 处理正文内容 (VitePress 通常使用 <main> 或 .vp-doc)
	c.OnHTML("main, .vp-doc, .theme-default-content", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] {
			return
		}
		visited[url] = true

		// 提取页面标题
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
			title = parts[len(parts)-1]
		}
		
		fmt.Printf("正在导出章节: %s\n", title)

		// 为代码块注入上下文注释，方便 AI 以后为你编写插件逻辑
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		htmlContent, _ := e.DOM.Html()
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			return
		}

		// 写入 Markdown 内容
		f.WriteString(fmt.Sprintf("# %s\n\n", title))
		f.WriteString(fmt.Sprintf("> 文档地址: [%s](%s)\n\n", url, url))
		f.WriteString(markdown)
		f.WriteString("\n\n\\newpage\n\n") // 强制 PDF 分页
	})

	// 4. 递归寻找侧边栏和正文链接
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// 只处理内部 Wiki 链接，排除锚点和外部跳转
		if strings.HasPrefix(link, basePath) || strings.HasPrefix(link, "/") {
			fullLink := e.Request.AbsoluteURL(link)
			if strings.Contains(fullLink, basePath) && !strings.Contains(fullLink, "#") {
				if !visited[fullLink] {
					e.Request.Visit(fullLink)
				}
			}
		}
	})

	fmt.Println("🚀 正在爬取 CraftEngine Wiki...")
	c.Visit("https://xiao-momi.github.io/craft-engine-wiki/")
	c.Wait()
	fmt.Println("✨ 知识库构建完成！")
}