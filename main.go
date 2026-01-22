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

type WikiSource struct {
	Name     string
	StartURL string
	BaseURL  string // 用于修复图片相对路径
	Domain   string
	Selector string
	Filter   string
}

func main() {
	combinedFile := "Minecraft_Dev_Encyclopedia.md"
	f, _ := os.Create(combinedFile)
	defer f.Close()

	// 写入百科全书元数据
	f.WriteString("---\n")
	f.WriteString("title: Minecraft 插件开发与运维百科全书\n")
	f.WriteString("author: 艾尔岚 (Ellan) 开发组\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	sources := []WikiSource{
		{"CustomCrops", "https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops", "https://mo-mi.gitbook.io", "mo-mi.gitbook.io", "main", "/customcrops"},
		{"JiuWu's Kitchen", "https://github.com/jiuwu02/JiuWu-s_Kitchen/wiki", "https://github.com", "github.com", "div.markdown-body", "/wiki"},
		{"CraftEngine", "https://xiao-momi.github.io/craft-engine-wiki/", "https://xiao-momi.github.io", "xiao-momi.github.io", "main, .vp-doc", "/craft-engine-wiki/"},
		{"The Brewing Project", "https://hangar.papermc.io/BreweryTeam/TheBrewingProject/pages/Wiki", "https://hangar.papermc.io", "hangar.papermc.io", ".project-page, .markdown-body", "/pages/"},
	}

	converter := md.NewConverter("", true, nil)

	for _, src := range sources {
		f.WriteString(fmt.Sprintf("\n\n# 📚 %s\n\n\\newpage\n", src.Name))
		visited := make(map[string]bool)
		c := colly.NewCollector(
			colly.AllowedDomains(src.Domain, "gitbook.io"),
			colly.UserAgent("Mozilla/5.0"),
		)

		c.OnHTML(src.Selector, func(e *colly.HTMLElement) {
			url := e.Request.URL.String()
			if visited[url] { return }
			visited[url] = true

			// 【关键修复】：修复图片相对路径
			e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
				imgSrc, exists := s.Attr("src")
				if exists && strings.HasPrefix(imgSrc, "/") {
					// 将 /assets/... 转换为 https://domain.com/assets/...
					s.SetAttr("src", src.BaseURL+imgSrc)
				}
			})

			// 标注代码块上下文 (对 EcoBridge 开发极其重要)
			e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
				s.PrependHtml(fmt.Sprintf("", src.Name))
			})

			html, _ := e.DOM.Html()
			markdown, _ := converter.ConvertString(html)
			f.WriteString(fmt.Sprintf("\n## %s\n\n%s\n\n\\newpage\n", e.DOM.Find("h1").First().Text(), markdown))
		})

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Request.AbsoluteURL(e.Attr("href"))
			if strings.Contains(link, src.Domain) && strings.Contains(link, src.Filter) && !strings.Contains(link, "#") {
				c.Visit(link)
			}
		})

		c.Visit(src.StartURL)
		c.Wait()
	}
}