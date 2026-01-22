package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Hangar V1 API 页面列表响应结构
type PageInfo struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PageContent struct {
	Markdown string `json:"markdown"`
}

func fetchAPI(client *http.Client, url string, target interface{}) error {
	// 确保 URL 去掉任何可能存在的空白字符或多余符号
	cleanURL := strings.TrimSpace(url)
	req, err := http.NewRequest("GET", cleanURL, nil)
	if err != nil {
		return err
	}

	// 必须包含 User-Agent 伪装，避免被 Cloudflare 拦截
	req.Header.Set("User-Agent", "EcoBridge-Doc-Bot/1.1 (GitHub Actions)")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
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
	f.WriteString("---\ntitle: The Brewing Project 官方 Wiki (V1 修复版)\nauthor: 艾尔岚开发组\ntoc: true\nlang: zh-CN\n---\n\n")

	// 核心参数：确保没有多余字符
	author := "BreweryTeam"
	project := "TheBrewingProject"
	baseURL := "https://hangar.papermc.io/api/v1"
	
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. 获取页面列表
	// Hangar V1 API 路径：/projects/{author}/{slug}/pages
	listURL := fmt.Sprintf("%s/projects/%s/%s/pages", baseURL, author, project)
	fmt.Printf("🚀 正在请求 API: %s\n", listURL)
	
	// 注意：Hangar V1 的 Pages 接口返回的是一个 Map[string]PageInfo
	var pagesMap map[string]PageInfo
	if err := fetchAPI(client, listURL, &pagesMap); err != nil {
		fmt.Printf("❌ 获取页面列表失败: %v\n", err)
		return
	}

	// 2. 遍历 Map 抓取内容
	for path, info := range pagesMap {
		fmt.Printf("📖 正在提取章节: %s (%s)\n", info.Name, path)
		
		contentURL := fmt.Sprintf("%s/projects/%s/%s/pages/%s", baseURL, author, project, path)
		var content PageContent
		if err := fetchAPI(client, contentURL, &content); err != nil {
			fmt.Printf("⚠️ 跳过页面 %s: %v\n", info.Name, err)
			continue
		}

		f.WriteString(fmt.Sprintf("# %s\n\n%s\n\n\\newpage\n\n", info.Name, content.Markdown))
		time.Sleep(300 * time.Millisecond) // 礼貌频率限制
	}

	fmt.Println("✨ 构建完成！")
}
