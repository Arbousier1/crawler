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
		// 稍微放宽域名限制，有时跳转会带上不同的后缀
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

		// 给代码块注入上下文
		e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
			s.PrependHtml(fmt.Sprintf("", title))
		})

		htmlContent, _ := e.DOM.Html()
		markdown, err := converter.ConvertString(htmlContent)
		if err != nil {
			return
		}

		// 注入 AI 识别标签
		annotated := strings.ReplaceAll(markdown, "```yaml", fmt.Sprintf("```yaml\n# Source: %s", url))
		final := fmt.Sprintf("# %s\n\n> URL: %s\n\n%s", title, url, annotated)

		fileName := sanitizeFilename(url) + ".md"
		os.WriteFile(filepath.Join(outputDir, fileName), []byte(final), 0644)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		// 关键：确保递归逻辑正确，只在 customcrops 路径下爬取
		if strings.Contains(link, "mo-mi.gitbook.io/xiaomomi-plugins/customcrops") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	fmt.Println("开始爬取 GitBook...")
	c.Visit("https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops")
	c.Wait()
}

func sanitizeFilename(url string) string {
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	hash := md5.Sum([]byte(url))
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(hash[:])[:4])
}
