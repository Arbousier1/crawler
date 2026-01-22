package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Hangar API 响应结构
type PageInfo struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PageContent struct {
	Markdown string `json:"markdown"`
}

func main() {
	outputFile := "TheBrewingProject_Wiki.md"
	f, _ := os.Create(outputFile)
	defer f.Close()

	// 写入元数据
	f.WriteString("---\n")
	f.WriteString("title: The Brewing Project 官方 Wiki (API 集成版)\n")
	f.WriteString("author: 艾尔岚 (Ellan) 开发助手\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	project := "BreweryTeam/TheBrewingProject"
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. 获取所有 Wiki 页面列表
	fmt.Println("正在从 API 获取页面列表...")
	listURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/projects/%s/pages", project)
	resp, err := client.Get(listURL)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("无法访问 API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var pages []PageInfo
	json.NewDecoder(resp.Body).Decode(&pages)

	// 2. 遍历并获取每个页面的原始 Markdown
	for _, page := range pages {
		fmt.Printf("正在抓取页面: %s\n", page.Name)
		
		// 构造页面内容的 API 链接
		contentURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/pages/page/%s/%s", project, page.Slug)
		cResp, cErr := client.Get(contentURL)
		if cErr != nil || cResp.StatusCode != 200 {
			continue
		}

		var content PageContent
		json.NewDecoder(cResp.Body).Decode(&content)
		cResp.Body.Close()

		// 3. 写入文件
		f.WriteString(fmt.Sprintf("# %s\n\n", page.Name))
		f.WriteString(content.Markdown)
		f.WriteString("\n\n\\newpage\n\n")
		
		// 稍微延迟，避免被 API 限制频率
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("✨ 抓取完成！已生成完整的 Markdown。")
}
