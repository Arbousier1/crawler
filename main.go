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
	Filter   string
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

	// 写入百科全书元数据
	f.WriteString("---\n")
	f.WriteString("title: Minecraft 开发百科全书 (技术版)\n")
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
			// 如果有英文版地址可在此更改 StartURL，目前保持根地址但通过逻辑过滤中文
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
			Selector: ".project-page, .markdown-body",
			Filter:   "/pages/",
		},
	}

	converter := md.NewConverter("", true, nil)

	for _, src := range sources {
		f.WriteString(fmt.Sprintf("\n\n# 📚 %s\n\n", src.Name))
		visited := make(map[string]bool)
		c := colly.NewCollector(
			colly.AllowedDomains(src.Domain, "gitbook.io"),
			colly.UserAgent("Mozilla/5.0"),
		)

		c.OnHTML(src.Selector, func(e *colly.HTMLElement) {
			url := e.Request.URL.String()

			// 【核心修改】：如果来源是 CraftEngine 且路径包含中文标识，则直接跳过
			if src.Name == "CraftEngine" && (strings.Contains(url, "/zh-Hans/") || strings.Contains(url, "/zh-CN/")) {
				return
			}

			if visited[url] { return }
			visited[url] = true

			fmt.Printf("[%s] 正在处理: %s\n", src.Name, url)

			// 修复图片路径
			e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
				imgSrc, _ := s.Attr("src")
				if strings.HasPrefix(imgSrc, "/") {
					s.SetAttr("src", src.BaseURL+imgSrc)
				}
			})

			html, _ := e.DOM.Html()
			markdown, _ := converter.ConvertString(html)
			cleanedMarkdown := cleanInternalLinks(markdown)

			f.WriteString(fmt.Sprintf("\n## %s\n\n%s\n\n\\newpage\n", e.DOM.Find("h1").First().Text(), cleanedMarkdown))
		})

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Request.AbsoluteURL(e.Attr("href"))

			// 【核心修改】：递归链接时也屏蔽中文路径
			if src.Name == "CraftEngine" && (strings.Contains(link, "/zh-Hans/") || strings.Contains(link, "/zh-CN/")) {
				return
			}

			if strings.Contains(link, src.Domain) && strings.Contains(link, src.Filter) && !strings.Contains(link, "#") {
				c.Visit(link)
			}
		})

		c.Visit(src.StartURL)
		c.Wait()
	}
}
