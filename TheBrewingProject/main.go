package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Hangar V1 API 页面列表响应是一个 Map
type PagesResponse map[string]struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PageContent struct {
	Markdown string `json:"markdown"`
}

func fetchWithHeader(client *http.Client, url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 必须包含 User-Agent 绕过 Cloudflare 基础校验
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) EcoBridge-Doc-Bot/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s (URL: %s)", resp.StatusCode, resp.Status, url)
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

	// 写入元数据
	f.WriteString("---\ntitle: The Brewing Project 官方 Wiki (V1 API 版)\nauthor: 艾尔岚开发组\ntoc: true\nlang: zh-CN\n---\n\n")

	// 官方 V1 API 路径
	author := "BreweryTeam"
	slug := "TheBrewingProject"
	baseURL := "https://hangar.papermc.io/api/v1"
	
	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("🚀 正在通过 V1 API 获取页面列表...")
	listURL := fmt.Sprintf("%s/projects/%s/%s/pages", baseURL, author, slug)
	
	var pagesMap PagesResponse
	if err := fetchWithHeader(client, listURL, &pagesMap); err != nil {
		fmt.Printf("❌ 获取页面列表失败: %v\n", err)
		return
	}

	// 遍历 Map 获取内容
	for path, info := range pagesMap {
		fmt.Printf("📖 正在提取章节: %s\n", info.Name)
		
		// V1 获取单个页面的接口
		contentURL := fmt.Sprintf("%s/projects/%s/%s/pages/%s", baseURL, author, slug, path)
		
		var content PageContent
		if err := fetchWithHeader(client, contentURL, &content); err != nil {
			fmt.Printf("⚠️ 跳过 %s: %v\n", info.Name, err)
			continue
		}

		f.WriteString(fmt.Sprintf("# %s\n\n%s\n\n\\newpage\n\n", info.Name, content.Markdown))
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println("✨ 构建完成！")
}
