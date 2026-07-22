package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
	"github.com/scagogogo/sploitus-skills/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// list 命令标志
	listSearchType string
	listSortType   string
	listOutputPath string
	listMaxResults int
)

var listCmd = &cobra.Command{
	Use:   "list [查询词]",
	Short: "获取所有查询结果",
	Long: `在Sploitus数据库中搜索并获取所有符合条件的漏洞利用信息，自动处理分页。
此命令会自动翻页获取所有匹配的结果。

示例:
  sploitus list "wordpress"
  sploitus list "cve:2021-44228" --type=cve
  sploitus list "apache" --sort=date --output=all_results.json
  sploitus list "sql injection" --max=100`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		// 创建客户端
		var results *types.SearchResponse
		var err error

		// 开始搜索
		if language == "en" {
			fmt.Printf("Searching: %q (type: %s, sort: %s)\n",
				query, listSearchType, listSortType)
			fmt.Println("Auto-fetching all results, please wait...")
		} else {
			fmt.Printf("正在搜索: %q (类型: %s, 排序: %s)\n",
				query, listSearchType, listSortType)
			fmt.Println("自动获取所有结果，请稍候...")
		}

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
			results, err = browser.Search(query, listSearchType, listSortType, 0)
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Browser search failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "浏览器搜索失败: %v\n", err)
				}
				os.Exit(1)
			}
		} else {
			client, err := createClient()
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				}
				os.Exit(1)
			}

			// 创建分页助手
			pagination := client.NewPaginationHelper(query, listSearchType, listSortType)

			// 获取第一页，了解总数
			results, err = pagination.GetFirstPage()
			if err != nil {
				if language == "en" {
					fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "搜索失败: %v\n", err)
				}
				os.Exit(1)
			}
		}

		totalItems := results.ExploitsTotal

		if useBrowser {
			// 浏览器模式下我们已经获取了所有结果
			if language == "en" {
				fmt.Printf("Found %d records\n", totalItems)
			} else {
				fmt.Printf("找到 %d 条记录\n", totalItems)
			}
		} else {
			// API模式下需要计算总页数
			client, _ := createClient()
			pagination := client.NewPaginationHelper(query, listSearchType, listSortType)
			totalPages, _ := pagination.GetTotalPages()

			if language == "en" {
				fmt.Printf("Found %d records (total %d pages)\n", totalItems, totalPages)
			} else {
				fmt.Printf("找到 %d 条记录 (共 %d 页)\n", totalItems, totalPages)
			}
		}

		if totalItems == 0 {
			if language == "en" {
				fmt.Println("No results found")
			} else {
				fmt.Println("未找到结果")
			}
			return
		}

		// 合并所有结果
		var allExploits []types.Exploit
		allExploits = append(allExploits, results.Exploits...)

		// 限制最大结果数
		limitedResults := totalItems
		if listMaxResults > 0 && listMaxResults < totalItems {
			limitedResults = listMaxResults
			if language == "en" {
				fmt.Printf("Will only fetch the first %d results\n", limitedResults)
			} else {
				fmt.Printf("将只获取前 %d 条结果\n", limitedResults)
			}
		}

		if !useBrowser {
			// 只有在API模式下才需要继续获取剩余页面
			// 已经获取了第一页，现在获取剩余页面
			if language == "en" {
				fmt.Print("Fetching results: ")
			} else {
				fmt.Print("正在获取结果: ")
			}
			currentPage := 1
			fmt.Printf("%d ", currentPage)

			client, _ := createClient()
			pagination := client.NewPaginationHelper(query, listSearchType, listSortType)

			for len(allExploits) < limitedResults && pagination.HasMore() {
				// 获取下一页
				nextPage, err := pagination.GetNextPage()
				if err != nil {
					if language == "en" {
						fmt.Fprintf(os.Stderr, "\nFailed to get next page: %v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "\n获取下一页失败: %v\n", err)
					}
					break
				}

				if len(nextPage.Exploits) == 0 {
					break
				}

				// 添加结果
				allExploits = append(allExploits, nextPage.Exploits...)

				// 检查是否达到限制
				if listMaxResults > 0 && len(allExploits) >= listMaxResults {
					allExploits = allExploits[:listMaxResults]
					break
				}

				// 打印进度
				currentPage++
				fmt.Printf("%d ", currentPage)
			}

			if language == "en" {
				fmt.Println("\nFetching completed")
			} else {
				fmt.Println("\n获取完成")
			}
		}

		// 创建完整结果
		completeResults := &types.SearchResponse{
			Exploits:      allExploits,
			ExploitsTotal: len(allExploits),
		}

		// 显示或保存结果
		if listOutputPath != "" {
			// 保存到文件
			saveToFile(completeResults, listOutputPath)
			if language == "en" {
				fmt.Printf("Saved %d results to %s\n", len(allExploits), listOutputPath)
			} else {
				fmt.Printf("已保存 %d 条结果到 %s\n", len(allExploits), listOutputPath)
			}
		} else {
			// 显示摘要
			if language == "en" {
				fmt.Printf("\nRetrieved %d result summaries:\n", len(allExploits))
				for i, exploit := range allExploits {
					fmt.Printf("%d. %s (Score: %.1f, Type: %s)\n",
						i+1, exploit.Title, exploit.Score, exploit.Type)
				}
				fmt.Println("\nUse the --output parameter to save full results to a file")
			} else {
				fmt.Printf("\n获取到 %d 条结果摘要:\n", len(allExploits))
				for i, exploit := range allExploits {
					fmt.Printf("%d. %s (Score: %.1f, Type: %s)\n",
						i+1, exploit.Title, exploit.Score, exploit.Type)
				}
				fmt.Println("\n使用 --output 参数可以将完整结果保存到文件")
			}
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&listSearchType, "type", "t", "", "搜索类型 (cve, title, tag)")
	listCmd.Flags().StringVarP(&listSortType, "sort", "s", "score", "排序方式 (score, date)")
	listCmd.Flags().StringVarP(&listOutputPath, "output", "o", "", "输出文件路径")
	listCmd.Flags().IntVarP(&listMaxResults, "max", "m", 0, "最大结果数量 (0表示不限制)")
	listCmd.Flags().BoolVarP(&useBrowser, "browser", "b", false, "使用浏览器自动化搜索")
	listCmd.Flags().BoolVarP(&debugBrowser, "debug", "d", false, "调试浏览器")
}

// 保存结果到文件
func saveToFile(results *types.SearchResponse, path string) error {
	var data []byte
	var err error

	if prettyPrint {
		data, err = json.MarshalIndent(results, "", "  ")
	} else {
		data, err = json.Marshal(results)
	}

	if err != nil {
		if language == "en" {
			return fmt.Errorf("JSON serialization failed: %w", err)
		}
		return fmt.Errorf("JSON序列化失败: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		if language == "en" {
			return fmt.Errorf("Failed to write to file: %w", err)
		}
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
