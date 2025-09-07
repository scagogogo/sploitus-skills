package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scagogogo/sploitus-crawler/pkg/sploitus"
	"github.com/scagogogo/sploitus-crawler/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// 搜索命令标志
	pageNumber      int
	pageSize        int
	searchType      string
	sortType        string
	searchOutputDir string
)

var searchCmd = &cobra.Command{
	Use:   "search [查询词]",
	Short: "搜索漏洞利用（支持分页）",
	Long: `搜索Sploitus数据库中的漏洞利用信息，支持分页。

示例:
  sploitus search "wordpress"
  sploitus search "cve:2021-44228" --type=cve
  sploitus search "apache" --sort=date --page=2 --size=20
  sploitus search "sql injection" --output=results.json`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		// 进行搜索
		if language == "en" {
			fmt.Printf("Searching for: %q (page: %d, size: %d)\n", query, pageNumber, pageSize)
		} else {
			fmt.Printf("搜索: %q (页码: %d, 每页结果数: %d)\n", query, pageNumber, pageSize)
		}

		var results *types.SearchResponse
		var err error

		if useBrowser {
			// 使用浏览器执行搜索
			if language == "en" {
				fmt.Println("Using browser automation to bypass CloudFlare protection...")
			} else {
				fmt.Println("正在使用自动化浏览器绕过CloudFlare保护...")
			}

			browser, browserErr := sploitus.NewBrowserSearcher(debugBrowser)
			if browserErr != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Failed to create browser: %v\n", browserErr)
				} else {
					fmt.Fprintf(os.Stderr, "创建浏览器失败: %v\n", browserErr)
				}
				os.Exit(1)
			}
			defer browser.Close()

			// 使用浏览器执行搜索
			offset := (pageNumber - 1) * pageSize
			results, err = browser.Search(query, searchType, sortType, offset)
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Browser search failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "浏览器搜索失败: %v\n", err)
				}
				os.Exit(1)
			}
		} else {
			// 使用常规API客户端
			client, err := createClient()
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				}
				os.Exit(1)
			}

			// 创建分页助手来执行搜索
			pagination := client.NewPaginationHelper(query, searchType, sortType)
			pagination.SetPageSize(pageSize)

			// 搜索特定页面
			if pageNumber == 1 {
				results, err = pagination.GetFirstPage()
			} else {
				results, err = pagination.GetPage(pageNumber)
			}

			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "搜索失败: %v\n", err)
				}
				os.Exit(1)
			}
		}

		// 处理输出
		if searchOutputDir != "" {
			// 保存到文件
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Failed to encode results: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "编码结果失败: %v\n", err)
				}
				os.Exit(1)
			}

			err = os.WriteFile(searchOutputDir, data, 0644)
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Failed to write results to file: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "写入结果到文件失败: %v\n", err)
				}
				os.Exit(1)
			}

			if language == "en" {
				fmt.Printf("Results saved to %s\n", searchOutputDir)
			} else {
				fmt.Printf("结果已保存到 %s\n", searchOutputDir)
			}
			return
		}

		// 输出到控制台
		if language == "en" {
			fmt.Printf("Found %d exploits (total: %d)\n\n", len(results.Exploits), results.ExploitsTotal)
		} else {
			fmt.Printf("找到 %d 条漏洞利用 (总共: %d)\n\n", len(results.Exploits), results.ExploitsTotal)
		}
		printResults(results.Exploits)

		// 显示分页信息
		if results.ExploitsTotal > pageSize {
			totalPages := (results.ExploitsTotal + pageSize - 1) / pageSize
			if language == "en" {
				fmt.Printf("\nPage %d of %d\n", pageNumber, totalPages)
				if pageNumber < totalPages {
					fmt.Printf("For next page, use: --page=%d\n", pageNumber+1)
				}
			} else {
				fmt.Printf("\n第 %d 页，共 %d 页\n", pageNumber, totalPages)
				if pageNumber < totalPages {
					fmt.Printf("查看下一页使用: --page=%d\n", pageNumber+1)
				}
			}
		}
	},
}

func init() {
	searchCmd.Flags().IntVarP(&pageNumber, "page", "g", 1, "页码")
	searchCmd.Flags().IntVarP(&pageSize, "size", "n", 10, "每页结果数")
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "搜索类型 (cve, title, tag)")
	searchCmd.Flags().StringVarP(&sortType, "sort", "s", "score", "排序方式 (score, date)")
	searchCmd.Flags().StringVarP(&searchOutputDir, "output", "o", "", "输出文件路径")
	searchCmd.Flags().BoolVarP(&useBrowser, "browser", "b", false, "使用浏览器自动化搜索")
	searchCmd.Flags().BoolVarP(&debugBrowser, "debug", "d", false, "调试浏览器")
}

// 打印搜索结果
func printResults(exploits []types.Exploit) {
	for i, exploit := range exploits {
		if language == "en" {
			fmt.Printf("%d. %s\n", i+1, exploit.Title)
			fmt.Printf("   ID: %s\n", exploit.ID)
			fmt.Printf("   Score: %.1f\n", exploit.Score)
			if exploit.Type != "" {
				fmt.Printf("   Type: %s\n", exploit.Type)
			}
			if exploit.Published != "" {
				fmt.Printf("   Published: %s\n", exploit.Published)
			}
			if exploit.Href != "" {
				fmt.Printf("   URL: %s\n", exploit.Href)
			}
		} else {
			fmt.Printf("%d. %s\n", i+1, exploit.Title)
			fmt.Printf("   ID: %s\n", exploit.ID)
			fmt.Printf("   得分: %.1f\n", exploit.Score)
			if exploit.Type != "" {
				fmt.Printf("   类型: %s\n", exploit.Type)
			}
			if exploit.Published != "" {
				fmt.Printf("   发布日期: %s\n", exploit.Published)
			}
			if exploit.Href != "" {
				fmt.Printf("   链接: %s\n", exploit.Href)
			}
		}
		fmt.Println()
	}
}
