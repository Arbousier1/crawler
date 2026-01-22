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

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gocolly/colly/v2"
)

func main() {
	// 1. 准备输出目录
	outputDir := "./knowledge_base"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	// 2. 初始化 Colly
	c := colly.NewCollector(
		colly.AllowedDomains("mo-mi.gitbook.io"),
		// 开启异步以提高速度，但需配合 Limit 使用
		colly.Async(true),
		colly.CacheDir("./colly_cache"),
	)

	// 限制并发，防止被 GitBook 拦截 (429 Too Many Requests)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 1 * time.Second,
	})

	// 初始化 Markdown 转换器
	converter := md.NewConverter("", true, nil)

	// 3. 处理内容 (GitBook 特定选择器)
	// GitBook 的主要内容通常在 <main> 标签中
	c.OnHTML("main", func(e *colly.HTMLElement) {
		url := e.Request.URL.String()
		
		// 过滤：只处理 CustomCrops 相关的页面，防止爬到该作者的其他插件文档
		if !strings.Contains(url, "/customcrops") {
			return
		}

		// 尝试获取标题，GitBook 标题通常是 main 下的第一个 h1
		title := e.DOM.Find("h1").First().Text()
		if title == "" {
			// 如果没找到 h1，尝试从 URL 获取最后一段作为标题
			parts := strings.Split(url, "/")
			if len(parts) > 0 {
				title = parts[len(parts)-1]
			} else {
				title = "Untitled"
			}
		}

		fmt.Printf("Crawling: %s -> %s\n", url, title)

		// 移除 GitBook 可能存在的 "Previous/Next" 底部导航链接，避免污染 AI 知识库
		e.DOM.Find("a[href*='/previous']").Remove()
		e.DOM.Find("a[href*='/next']").Remove()

		// 转换为 Markdown
		markdown, err := converter.ConvertString(e.HTML)
		if err != nil {
			log.Printf("Error converting %s: %v", url, err)
			return
		}

		// 构建 Frontmatter (元数据)
		// 这对 RAG 很重要，因为它告诉 AI 这个知识的来源
		finalContent := fmt.Sprintf("---\nsource_url: %s\ntitle: %s\ncrawled_at: %s\n---\n\n%s",
			url, title, time.Now().Format("2006-01-02"), markdown)

		fileName := sanitizeFilename(url) + ".md"
		filePath := filepath.Join(outputDir, fileName)

		if err := os.WriteFile(filePath, []byte(finalContent), 0644); err != nil {
			log.Printf("Error writing file %s: %v", filePath, err)
		}
	})

	// 4. 链接发现 (递归爬取)
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		
		// 确保只访问该文档库内部的链接
		// 必须包含 /xiaomomi-plugins/ 且不包含 # (锚点)
		if strings.Contains(link, "/xiaomomi-plugins/") && !strings.Contains(link, "#") {
			e.Request.Visit(link)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Failed: %s (Status: %d) - %v", r.Request.URL, r.StatusCode, err)
	})

	fmt.Println("Starting GitBook crawler...")
	// 入口链接
	c.Visit("https://mo-mi.gitbook.io/xiaomomi-plugins/customcrops")
	
	// 等待所有异步任务完成
	c.Wait()
}

// 保持不变的文件名清理函数
func sanitizeFilename(url string) string {
	hash := md5.Sum([]byte(url))
	shortHash := hex.EncodeToString(hash[:])[:6]
	
	cleanName := strings.ReplaceAll(url, "https://", "")
	cleanName = strings.ReplaceAll(cleanName, "mo-mi.gitbook.io/", "")
	cleanName = strings.ReplaceAll(cleanName, "xiaomomi-plugins/", "")
	cleanName = strings.ReplaceAll(cleanName, "/", "-")
	
	if len(cleanName) > 60 {
		cleanName = cleanName[:60] + "-" + shortHash
	}
	return cleanName
}
