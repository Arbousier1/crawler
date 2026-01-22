package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PageInfo struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PageContent struct {
	Markdown string `json:"markdown"`
}

// 通用的请求函数，包含必要的 Header 伪装
func fetch(client *http.Client, url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 必须设置 User-Agent，否则 Hangar 会返回 403 错误
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 状态异常: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func main() {
	outputFile := "TheBrewingProject_Wiki.md"
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("❌ 无法创建文件: %v\n", err)
		return
	}
	defer f.Close()

	f.WriteString("---\ntitle: The Brewing Project 官方 Wiki (API 版)\nauthor: 自动化助理\ntoc: true\nlang: zh-CN\n---\n\n")

	project := "BreweryTeam/TheBrewingProject"
	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("🚀 正在从 Hangar API 获取页面索引...")
	listURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/projects/%s/pages", project)
	
	var pages []PageInfo
	if err := fetch(client, listURL, &pages); err != nil {
		fmt.Printf("❌ 获取页面列表失败: %v\n", err)
		return
	}

	for _, page := range pages {
		fmt.Printf("📖 正在提取页面: %s\n", page.Name)
		contentURL := fmt.Sprintf("https://hangar.papermc.io/api/internal/pages/page/%s/%s", project, page.Slug)
		
		var content PageContent
		if err := fetch(client, contentURL, &content); err != nil {
			fmt.Printf("⚠️ 跳过页面 %s: %v\n", page.Name, err)
			continue
		}

		f.WriteString(fmt.Sprintf("# %s\n\n%s\n\n\\newpage\n\n", page.Name, content.Markdown))
		time.Sleep(300 * time.Millisecond) // 避免请求过快
	}

	fmt.Println("✨ 构建完成！")
}
