package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Hangar V1 页面响应结构
type PageInfo struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PageContent struct {
	Markdown string `json:"markdown"`
}

func fetchHangar(client *http.Client, url string, target interface{}) error {
	// 彻底清理 URL，确保没有多余的空格或括号
	cleanURL := strings.TrimSpace(url)
	req, err := http.NewRequest("GET", cleanURL, nil)
	if err != nil {
		return err
	}

	// 遵循官方准则：设置有意义的 User-Agent
	req.Header.Set("User-Agent", "EcoBridge-Knowledge-Base-Bot/1.0 (Contact: Ellan-Dev-Group)")
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

	// 写入合规元数据
	f.WriteString("---\ntitle: The Brewing Project 官方百科 (V1 API)\nauthor: 艾尔岚开发组\ntoc: true\nlang: zh-CN\n---\n\n")

	// 官方参数
	author := "BreweryTeam"
	project := "TheBrewingProject"
	baseURL := "https://hangar.papermc.io/api/v1"
	
	// 官方建议 Anonymous 访问公开信息不需要 JWT
	client := &http.Client{Timeout: 30 * time.Second}

	fmt.Println("🚀 正在遵循 Hangar V1 准则获取页面列表...")
	listURL := fmt.Sprintf("%s/projects/%s/%s/pages", baseURL, author, project)
	
	var pagesMap map[string]PageInfo
	if err := fetchHangar(client, listURL, &pagesMap); err != nil {
		fmt.Printf("❌ API 访问失败: %v\n", err)
		return
	}

	// 遵循速率限制：官方默认 20req/5s
	// 我们设置每个请求间隔 300ms 确保绝对安全
	for path, info := range pagesMap {
		fmt.Printf("📖 正在同步章节: %s\n", info.Name)
		
		contentURL := fmt.Sprintf("%s/projects/%s/%s/pages/%s", baseURL, author, project, path)
		var content PageContent
		if err := fetchHangar(client, contentURL, &content); err != nil {
			fmt.Printf("⚠️ 页面获取中断: %v\n", err)
			continue
		}

		f.WriteString(fmt.Sprintf("# %s\n\n%s\n\n\\newpage\n\n", info.Name, content.Markdown))
		time.Sleep(300 * time.Millisecond) 
	}

	fmt.Println("✨ 抓取任务圆满完成！")
}
