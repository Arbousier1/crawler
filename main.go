package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

func main() {
	outputDir := "./knowledge_base"
	os.MkdirAll(outputDir, 0755)

	visited := make(map[string]bool)
	c := colly.NewCollector(
		colly.AllowedDomains("mo-mi.gitbook.io", "gitbook.io"),
		colly.Async(true),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 1 * time.Second,
	})

	converter := md.NewConverter("", true, nil)

	c.OnHTML("main", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] || !strings.Contains(url, "customcrops") {
			return
		}
		visited[url] = true

		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			title = "CustomCrops Doc"
		}
		
		fmt.Printf("成功发现页面: %s\n", title)

		// 注入上下文注释到 HTML 的 pre 标签中
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title, url))
		})

		htmlContent, err := e.DOM.Html()
		if err != nil {
			return
		}

		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			return
		}

		// 增强：给 Markdown 代码块块首添加语义化注释
		annotated := strings.ReplaceAll(markdown, "```yaml", fmt.Sprintf("```yaml\n# 来自文档: %s\n# 原始链接: %s", title, url))
		annotated = strings.ReplaceAll(annotated, "```yml", fmt.Sprintf("```yml\n# 来自文档: %s\n# 原始链接: %s", title, url))

		final := fmt.Sprintf("# %s\n\n> URL: %s\n> Exported: %s\n\n---\n\n%s", 
			title, url, time.Now().Format("2006-01-02"), annotated)

		fileName := sanitizeFilename(url) + ".md"
		os.WriteFile(filepath.Join(outputDir, fileName), []byte(final), 0644)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Contains(link, "mo-mi.gitbook.io/xiaomomi-plugins/customcrops") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	fmt.Println("🚀 开始爬取 GitBook 并构建 AI 知识库...")
	c.Visit("https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops")
	c.Wait()
	fmt.Println("✨ 导出完成！")
}

func sanitizeFilename(url string) string {
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	if name == "customcrops" || name == "" {
		name = "home"
	}
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(hash[:])[:4])
}
