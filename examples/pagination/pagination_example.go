package main

import (
	"fmt"
	"log"

	"github.com/scagogogo/sploitus-crawler/pkg/sploitus"
	"github.com/scagogogo/sploitus-crawler/pkg/types"
)

func main() {
	// 创建客户端
	client := sploitus.NewClient()

	// 创建分页助手（默认中文版）
	paginator := client.NewPaginationHelper("CVE-2023", "exploits", "default")

	// 设置每页显示10条结果
	paginator.SetPageSize(10)

	// 获取第一页结果
	firstPage, err := paginator.GetFirstPage()
	if err != nil {
		log.Fatalf("获取第一页失败: %v", err)
	}

	// 显示结果信息
	fmt.Println("===== 第一页 =====")
	fmt.Println(paginator.GetPageInfo())
	fmt.Printf("结果数量: %d\n", len(firstPage.Exploits))
	for i, exploit := range firstPage.Exploits {
		fmt.Printf("%d. %s (评分: %.1f)\n", i+1, exploit.Title, exploit.Score)
	}

	// 获取下一页
	if paginator.HasMore() {
		nextPage, err := paginator.GetNextPage()
		if err != nil {
			log.Fatalf("获取下一页失败: %v", err)
		}

		fmt.Println("\n===== 第二页 =====")
		fmt.Println(paginator.GetPageInfo())
		fmt.Printf("结果数量: %d\n", len(nextPage.Exploits))
		for i, exploit := range nextPage.Exploits {
			fmt.Printf("%d. %s (评分: %.1f)\n", i+1, exploit.Title, exploit.Score)
		}
	}

	// 获取特定页码
	page3, err := paginator.GetPage(3)
	if err != nil {
		log.Fatalf("获取第3页失败: %v", err)
	}

	fmt.Println("\n===== 第三页 =====")
	fmt.Println(paginator.GetPageInfo())
	fmt.Printf("结果数量: %d\n", len(page3.Exploits))

	// 获取总页数
	totalPages, err := paginator.GetTotalPages()
	if err != nil {
		log.Fatalf("获取总页数失败: %v", err)
	}
	fmt.Printf("\n总共有 %d 页结果\n", totalPages)

	// 使用英文版分页器
	fmt.Println("\n===== English Pagination =====")
	enPaginator := client.NewEnPaginationHelper("CVE-2023", "exploits", "default")
	enPaginator.SetPageSize(10)

	enFirstPage, err := enPaginator.GetFirstPage()
	if err != nil {
		log.Fatalf("Failed to get first page: %v", err)
	}

	fmt.Println(enPaginator.GetPageInfo())
	fmt.Printf("Number of results: %d\n", len(enFirstPage.Exploits))

	// 高级用法：一次性获取所有结果（谨慎使用，可能会有很多数据）
	fmt.Println("\n===== 获取部分数据示例 =====")
	// 创建新的分页器，设置较小的页面大小以便演示
	demoPageSize := 5
	paginator = client.NewPaginationHelper("CVE-2023-1234", "exploits", "default")
	paginator.SetPageSize(demoPageSize)

	// 模拟浏览前3页
	var allResults []string
	for i := 0; i < 3; i++ {
		if !paginator.HasMore() {
			break
		}

		var page *types.SearchResponse
		if i == 0 {
			page, err = paginator.GetFirstPage()
		} else {
			page, err = paginator.GetNextPage()
		}

		if err != nil {
			log.Fatalf("获取页面失败: %v", err)
		}

		fmt.Printf("--- 页面 %d ---\n", paginator.GetCurrentPage())

		if len(page.Exploits) == 0 {
			fmt.Println("没有更多结果")
			break
		}

		for _, exploit := range page.Exploits {
			allResults = append(allResults, exploit.Title)
		}
	}

	fmt.Printf("\n收集到 %d 条结果\n", len(allResults))
}
