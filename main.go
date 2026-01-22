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

type WikiSource struct {
	Name     string
	StartURL string
	BaseURL  string
	Domain   string
	Selector string
	Filter   string // 必须包含的路径关键词
}

// 清理无效内部锚点链接，解决 Pandoc 转换 PDF 时的报错
func cleanInternalLinks(content string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(#[^\)]+\)`)
	return re.ReplaceAllString(content, "$1")
}

func main() {
	combinedFile := "Minecraft_Dev_Encyclopedia.md"
	f, err := os.Create(combinedFile)
	if err != nil {
		fmt.Printf("无法创建文件: %v\n", err)
		return
	}
	defer f.Close()

	f.WriteString("---\n")
	f.WriteString("title: Minecraft 开发百科全书 (全集成版)\n")
	f.WriteString("author: 艾尔岚 (Ellan) 开发组\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	sources := []WikiSource{
		{
			Name:     "CustomCrops",
			StartURL: "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops",
			BaseURL:  "https://mo-mi.gitbook.io",
			Domain:   "mo-mi.gitbook.io",
			Selector: "main",
			Filter:   "/customcrops",
		},
		{
			Name:     "JiuWu's Kitchen",
			StartURL: "https://github.com/jiuwu02/JiuWu-s_Kitchen/wiki",
			BaseURL:  "https://github.com",
			Domain:   "github.com",
			Selector: "div.markdown-body",
			Filter:   "/wiki",
		},
		{
			Name:     "CraftEngine",
			StartURL: "https://xiao-momi.github.io/craft-engine-wiki/",
			BaseURL:  "https://xiao-momi.github.io",
			Domain:   "xiao-momi.github.io",
			Selector: "main, .vp-doc",
			Filter:   "/craft-engine-wiki/",
		},
		{
			Name:     "The Brewing Project",
			StartURL: "https://hangar.papermc.io/BreweryTeam/TheBrewingProject/pages/Wiki",
			BaseURL:  "https://hangar.papermc.io",
			Domain:   "hangar.papermc.io",
			// Hangar 的正文通常在 .project-page 或 markdown-body 内
			Selector: ".project-page, .markdown-content, .markdown-body",
			// 修正 Filter 以匹配 Hangar 的子页面路径
			Filter:   "/BreweryTeam/TheBrewingProject/pages/",
		},
	}

	converter := md.NewConverter("", true, nil)

	for _, src := range sources {
		f.WriteString(fmt.Sprintf("\n\n# 📚 %s\n\n", src.Name))
		visited := make(map[string]bool)
		c := colly.NewCollector(
			colly.AllowedDomains(src.Domain, "gitbook.io"),
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			colly.Async(true),
		)

		c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2, RandomDelay: 1 * time.Second})

		c.OnHTML(src.Selector, func(e *colly.HTMLElement) {
			url := e.Request.URL.String()

			// 1. 屏蔽 CraftEngine 中文 Wiki
			if src.Name == "CraftEngine" && (strings.Contains(url, "/zh-Hans/") || strings.Contains(url, "/zh-CN/")) {
				return
			}

			if visited[url] { return }
			visited[url] = true

			title := e.DOM.Find("h1").First().Text()
			if title == "" {
				parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
				title = parts[len(parts)-1]
			}
			fmt.Printf("[%s] 正在处理: %s\n", src.Name, title)

			// 2. 修复图片路径 (WeasyPrint 转换必需)
			e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
				imgSrc, _ := s.Attr("src")
				if strings.HasPrefix(imgSrc, "/") {
					s.SetAttr("src", src.BaseURL+imgSrc)
				}
			})

			// 3. 标注上下文，方便 EcoBridge 开发时 AI 识别
			e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
				s.PrependHtml(fmt.Sprintf("", src.Name, title))
			})

			html, _ := e.DOM.Html()
			markdown, _ := converter.ConvertString(html)
			cleanedMarkdown := cleanInternalLinks(markdown)

			f.WriteString(fmt.Sprintf("\n## %s\n\n%s\n\n\\newpage\n", title, cleanedMarkdown))
		})

		// 4. 递归寻找文档链接
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Request.AbsoluteURL(e.Attr("href"))
			
			// 屏蔽 CraftEngine 中文链接递归
			if src.Name == "CraftEngine" && (strings.Contains(link, "/zh-Hans/") || strings.Contains(link, "/zh-CN/")) {
				return
			}

			// 适配 Hangar 的路径识别
			if strings.Contains(link, src.Domain) && strings.Contains(link, src.Filter) && !strings.Contains(link, "#") {
				if !visited[link] {
					e.Request.Visit(link)
				}
			}
		})

		c.Visit(src.StartURL)
		c.Wait()
	}
	fmt.Println("✨ 跨平台百科全书构建完成！")
}
