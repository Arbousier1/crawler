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
	combinedFile := "kitchen_wiki_full.md"
	f, err := os.Create(combinedFile)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer f.Close()

	// 2. 写入 PDF 元数据头部
	f.WriteString("---\n")
	f.WriteString("title: JiuWu's Kitchen 完整指南\n")
	f.WriteString("author: AI 知识库同步助手\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("toc-title: 目录\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	visited := make(map[string]bool)
	// GitHub Wiki 的基础路径过滤
	wikiPath := "/jiuwu02/JiuWu-s_Kitchen/wiki"

	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	converter := md.NewConverter("", true, nil)

	// 3. 处理 Wiki 正文内容
	c.OnHTML("div.markdown-body", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] {
			return
		}
		visited[url] = true

		// 提取标题
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			parts := strings.Split(url, "/")
			title = parts[len(parts)-1]
		}
		
		fmt.Printf("正在提取页面: %s\n", title)

		// 为代码块注入上下文，方便 AI 识别
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		htmlContent, _ := e.DOM.Html()
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			return
		}

		// 增强：给 Markdown 代码块块首添加语义化注释
		annotated := strings.ReplaceAll(markdown, "```yaml", fmt.Sprintf("```yaml\n# 来自文档: %s", title))
		annotated = strings.ReplaceAll(annotated, "```yml", fmt.Sprintf("```yml\n# 来自文档: %s", title))

		// 写入内容
		f.WriteString(fmt.Sprintf("# %s\n\n", title))
		f.WriteString(fmt.Sprintf("> 原始链接: [%s](%s)\n\n", url, url))
		f.WriteString(annotated)
		f.WriteString("\n\n\\newpage\n\n")
	})

	// 4. 递归寻找 Wiki 侧边栏及正文中的其他页面链接
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// 只抓取本仓库 Wiki 范围内的链接，排除锚点
		if strings.HasPrefix(link, wikiPath) && !strings.Contains(link, "#") {
			fullLink := e.Request.AbsoluteURL(link)
			if !visited[fullLink] {
				e.Request.Visit(fullLink)
			}
		}
	})

	fmt.Println("🚀 开始爬取 GitHub Wiki...")
	c.Visit("https://github.com/jiuwu02/JiuWu-s_Kitchen/wiki")
	c.Wait()
	fmt.Println("✨ 文档集成完成！")
}