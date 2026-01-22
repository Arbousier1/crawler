package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

func main() {
	// 最终合并的 Markdown 文件
	combinedFile := "customcrops_wiki.md"
	f, _ := os.Create(combinedFile)
	defer f.Close()

	// 写入 PDF 元数据和标题
	f.WriteString(fmt.Sprintf("%% CustomCrops 完整插件文档\n%% 自动构建机器人\n%% 更新日期: %s\n\n", time.Now().Format("2006-01-02")))

	visited := make(map[string]bool)
	c := colly.NewCollector(
		colly.AllowedDomains("mo-mi.gitbook.io", "gitbook.io"),
		colly.UserAgent("Mozilla/5.0"),
	)

	converter := md.NewConverter("", true, nil)

	c.OnHTML("main", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] || !strings.Contains(url, "customcrops") {
			return
		}
		visited[url] = true

		title := e.DOM.Find("h1").First().Text()
		fmt.Printf("正在抓取页面: %s\n", title)

		// 标注代码块上下文
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		htmlContent, _ := e.DOM.Html()
		markdown, _ := converter.ConvertString(htmlContent)

		// 格式化：每个页面作为二级标题，并强制代码块标注
		annotated := strings.ReplaceAll(markdown, "```yaml", fmt.Sprintf("```yaml\n# 来自文档: %s", title))
		
		// 写入合并文件
		f.WriteString(fmt.Sprintf("\n\n# %s\n\n> 原始链接: %s\n\n%s\n\n\\newpage\n", title, url, annotated))
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Contains(link, "mo-mi.gitbook.io/xiaomomi-plugins/customcrops") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	c.Visit("https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops")
	c.Wait()
}
