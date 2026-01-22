package main

import (
	"encoding/json"
	"fmt"
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
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("无法创建文件: %v\n", err)
		return
	}
	defer f.Close()

	// 写入元数据 (Pandoc 兼容)
	f.WriteString("---\n")
	f.WriteString("title: The Brewing Project 官方 Wiki (API 版)\n")
	f.WriteString("author: 艾尔岚开发组\n")
	f.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))
	f.WriteString("toc: true\n")
	f.WriteString("lang: zh-CN\n")
	f.WriteString("---\n\n")

	project := "BreweryTeam/TheBrewingProject"
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. 获取所有 Wiki 页面列表
	fmt.Println("🚀 正在从 Hangar API 获取页面索引...")
	listURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/projects/%s/pages", project)
	resp, err := client.Get(listURL)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("API 访问失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var pages []PageInfo
	if err := json.NewDecoder(resp.Body).Decode(&pages); err != nil {
		fmt.Printf("解析 JSON 失败: %v\n", err)
		return
	}

	// 2. 遍历并拉取原始 Markdown
	for _, page := range pages {
		fmt.Printf("正在提取页面: %s\n", page.Name)
		
		contentURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/pages/page/%s/%s", project, page.Slug)
		cResp, cErr := client.Get(contentURL)
		if cErr != nil || cResp.StatusCode != 200 {
			continue
		}

		var content PageContent
		json.NewDecoder(cResp.Body).Decode(&content)
		cResp.Body.Close()

		// 写入 Markdown
		f.WriteString(fmt.Sprintf("# %s\n\n", page.Name))
		f.WriteString(content.Markdown)
		f.WriteString("\n\n\\newpage\n\n")
		
		time.Sleep(200 * time.Millisecond) // 适度延迟
	}

	fmt.Println("✨ 抓取完成！百科全书 Markdown 已就绪。")
}
