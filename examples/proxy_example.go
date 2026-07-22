package main

import (
	"fmt"
	"log"

	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
)

func main() {
	// 方法1：创建带有代理的客户端
	proxyURL := "http://localhost:8080" // 更改为您的代理服务器地址
	client, err := sploitus.NewClientWithProxy(proxyURL)
	if err != nil {
		log.Fatalf("Failed to create client with proxy: %v", err)
	}

	// 使用代理进行搜索
	results, err := client.Search("CVE-2023-12345", "exploits", "default", 0)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	fmt.Printf("Found %d exploits using proxy\n", results.ExploitsTotal)

	// 方法2：创建普通客户端后设置代理
	regularClient := sploitus.NewClient()

	// 之后可以随时设置或更改代理
	err = regularClient.SetProxy(proxyURL)
	if err != nil {
		log.Fatalf("Failed to set proxy: %v", err)
	}

	// 也可以使用不同的代理
	// err = regularClient.SetProxy("http://other-proxy:8080")

	// 使用设置了代理的客户端进行搜索
	results, err = regularClient.Search("CVE-2023-67890", "exploits", "default", 0)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}
	fmt.Printf("Found %d exploits after setting proxy\n", results.ExploitsTotal)
}
