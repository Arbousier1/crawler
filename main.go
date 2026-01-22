package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
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
		colly.AllowedDomains("mo-mi.gitbook.io"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 3,
		RandomDelay: 1 * time.Second,
	})

	// 1. 初始化转换器
	converter := md.NewConverter("", true, nil)

	// 2. 【核心增强】：添加自定义规则来标注代码块
	converter.AddRules(
		md.Rule{
			Filter: []string{"pre"},
			Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
				// 获取当前页面的上下文信息（从 selection 的 parent 向上找或者通过全局变量，但在 Colly 中我们直接处理 HTML）
				// 这里我们简单地通过 content 处理，但更优雅的是在 OnHTML 中处理。
				// 由于 html-to-markdown 是流式处理，我们通过下面的注入方式实现。
				return nil // 返回 nil 表示使用默认处理
			},
		},
	)

	c.OnHTML("main", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		if visited[url] {
			return
		}
		visited[url] = true

		if !strings.Contains(url, "/customcrops") {
			return
		}

		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			title = "Untitled"
		}
		fmt.Printf("[Parsing] %s\n", title)

		// 3. 【精准增强】：在转换前，手动给 HTML 里的代码块加注
		// 这样 AI 在看到代码时，第一行永远是上下文注释
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			contextInfo := fmt.Sprintf("\n# Context: %s (Source: %s)\n", title, url)
			s.PrependHtml(fmt.Sprintf("", contextInfo)) 
			// 或者直接在代码块上方插入一个段落
		})

		htmlContent, _ := e.DOM.Html()
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			log.Printf("Conversion failed: %v", err)
			return
		}

		// 4. 在代码块中强制注入语义化注释
		// 这一步利用字符串替换，把 Markdown 的 ```yaml 替换为带注释的版本
		annotatedMarkdown := strings.ReplaceAll(markdown, "```yaml", fmt.Sprintf("```yaml\n# 来自文档: %s\n# 路径: %s", title, url))
		annotatedMarkdown = strings.ReplaceAll(annotatedMarkdown, "```yml", fmt.Sprintf("```yml\n# 来自文档: %s\n# 路径: %s", title, url))

		finalContent := fmt.Sprintf("# %s\n\n> URL: %s\n\n---\n\n%s", title, url, annotatedMarkdown)

		fileName := sanitizeFilename(url) + ".md"
		os.WriteFile(filepath.Join(outputDir, fileName), []byte(finalContent), 0644)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if strings.Contains(link, "mo-mi.gitbook.io/xiaomomi-plugins/customcrops") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	c.Visit("[https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops](https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops)")
	c.Wait()
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
