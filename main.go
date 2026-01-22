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

// WikiSource 定义不同文档源的配置
type WikiSource struct {
	Name     string
	StartURL string
	BaseURL  string
	Domain   string
	Selector string
	Filter   string // 链接递归过滤器
}

// cleanInternalLinks 核心修复逻辑：移除形如 [#anchor] 的内部死链，保留文本
// 这能解决 Pandoc 转换 PDF 时报 "No anchor for internal URI reference" 的错误
func cleanInternalLinks(content string) string {
	// 匹配 [文字](#锚点) 并替换为 [文字] 或直接替换为 文字
	re := regexp.MustCompile(`\[([^\]]+)\]\(#[^\)]+\)`)
	return re.ReplaceAllString(content, "$1")
}

func main() {
	combinedFile := "Minecraft_Dev_Encyclopedia.md"
	
	// 1. 创建并初始化合并文件
	f, err := os.Create(combinedFile)
	if err != nil {
		fmt.Printf("无法创建文件: %v\n", err)
		return
	}
	defer f.Close()

	// 写入 Pandoc 兼容的 YAML 元数据
	f.WriteString("---\n")
	f.WriteString("title: Minecraft 插件开发与运维百科全书\n")
	f.WriteString("author: 艾尔岚 (Ellan) 开发组\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("toc-title: 百科全书目录\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("breakurl: true\n")
	f.WriteString("colorlinks: true\n")
	f.WriteString("---\n\n")

	// 2. 定义四个核心文档源
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
			Selector: ".project-page, .markdown-body",
			Filter:   "/pages/",
		},
	}

	converter := md.NewConverter("", true, nil)

	// 3. 循环爬取每一个源
	for _, src := range sources {
		f.WriteString(fmt.Sprintf("\n\n# 📚 插件大类：%s\n\n\\newpage\n", src.Name))
		
		visited := make(map[string]bool)
		c := colly.NewCollector(
			colly.AllowedDomains(src.Domain, "gitbook.io"),
			colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			colly.Async(true),
		)

		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 2,
			RandomDelay: 1 * time.Second,
		})

		// 处理 HTML 并转换为 Markdown
		c.OnHTML(src.Selector, func(e *colly.HTMLElement) {
			url := e.Request.URL.String()
			if visited[url] {
				return
			}
			visited[url] = true

			// 获取页面标题
			pageTitle := e.DOM.Find("h1").First().Text()
			if pageTitle == "" {
				parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
				pageTitle = parts[len(parts)-1]
			}
			fmt.Printf("[%s] 正在处理页面: %s\n", src.Name, pageTitle)

			// 核心修复：将图片相对路径补全为绝对路径
			e.DOM.Find("img").Each(func(i int, s *goquery.Selection) {
				imgSrc, exists := s.Attr("src")
				if exists && strings.HasPrefix(imgSrc, "/") {
					s.SetAttr("src", src.BaseURL+imgSrc)
				}
			})

			// AI 辅助增强：在代码块中注入项目上下文，方便 EcoBridge 开发时识别
			e.DOM.Find("pre").Each(func(i int, s *goquery.Selection) {
				s.PrependHtml(fmt.Sprintf("\n", src.Name))
			})

			html, _ := e.DOM.Html()
			markdown, err := converter.ConvertString(html)
			if err != nil {
				return
			}

			// 核心修复：清理会导致 PDF 报错的内部死链
			cleanedMarkdown := cleanInternalLinks(markdown)

			// 写入合并文件
			f.WriteString(fmt.Sprintf("\n## [%s] %s\n\n", src.Name, pageTitle))
			f.WriteString(fmt.Sprintf("> 来源: %s\n\n", url))
			f.WriteString(cleanedMarkdown)
			f.WriteString("\n\n\\newpage\n\n")
		})

		// 递归发现链接
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Request.AbsoluteURL(e.Attr("href"))
			// 仅在本插件 Wiki 路径内递归，防止爬虫逃逸
			if strings.Contains(link, src.Domain) && strings.Contains(link, src.Filter) && !strings.Contains(link, "#") {
				c.Visit(link)
			}
		})

		c.Visit(src.StartURL)
		c.Wait()
	}

	fmt.Println("✨ 百科全书 Markdown 构建完成！请运行 Pandoc 转换为 PDF。")
}
