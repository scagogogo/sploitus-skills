package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scagogogo/sploitus-crawler/pkg/types"
	"github.com/spf13/cobra"
)

var (
	// payload 命令标志
	payloadSearchType string
	payloadSortType   string
	payloadMaxResults int
	payloadOutputDir  string
	payloadNaming     string // 命名格式: id (CVE ID), title (标题), both (两者)
	payloadLang       string // 输出语言: cn (中文), en (英文)
)

var payloadCmd = &cobra.Command{
	Use:   "payload [查询词]",
	Short: "搜索并保存漏洞利用代码",
	Long: `根据关键词搜索漏洞利用代码，并将结果保存到本地文件夹。
此命令会自动获取所有匹配结果，并为每个结果创建单独的文件，文件格式会根据漏洞利用的语言自动选择。
结果元数据（如标题、得分、链接等）会作为注释添加在文件头部。

示例:
  sploitus payload "wordpress"
  sploitus payload "cve:2021-44228" --type=cve --output=./exploits
  sploitus payload "sql injection" --max=50 --naming=title
  sploitus payload "log4j" --sort=date --naming=both --lang=en`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		// 创建客户端
		client, err := createClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		// 创建分页助手
		pagination := client.NewPaginationHelper(query, payloadSearchType, payloadSortType)

		// 开始搜索
		if payloadLang == "en" {
			fmt.Printf("Searching for exploit code: %q (type: %s, sort: %s)\n",
				query, payloadSearchType, payloadSortType)
			fmt.Println("Fetching matching results, please wait...")
		} else {
			fmt.Printf("正在搜索漏洞利用代码: %q (类型: %s, 排序: %s)\n",
				query, payloadSearchType, payloadSortType)
			fmt.Println("自动获取匹配结果，请稍候...")
		}

		// 获取第一页，了解总数
		firstPage, err := pagination.GetFirstPage()
		if err != nil {
			if payloadLang == "en" {
				fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "搜索失败: %v\n", err)
			}
			os.Exit(1)
		}

		totalItems := firstPage.ExploitsTotal
		totalPages, _ := pagination.GetTotalPages()

		if payloadLang == "en" {
			fmt.Printf("Found %d exploit(s) (total %d pages)\n", totalItems, totalPages)
		} else {
			fmt.Printf("找到 %d 条漏洞利用 (共 %d 页)\n", totalItems, totalPages)
		}

		if totalItems == 0 {
			if payloadLang == "en" {
				fmt.Println("No results found")
			} else {
				fmt.Println("未找到结果")
			}
			return
		}

		// 限制最大结果数
		limitedResults := totalItems
		if payloadMaxResults > 0 && payloadMaxResults < totalItems {
			limitedResults = payloadMaxResults
			if payloadLang == "en" {
				fmt.Printf("Will only fetch the first %d results\n", limitedResults)
			} else {
				fmt.Printf("将只获取前 %d 条结果\n", limitedResults)
			}
		}

		// 合并所有结果
		var allExploits []types.Exploit
		allExploits = append(allExploits, firstPage.Exploits...)

		// 已经获取了第一页，现在获取剩余页面
		if payloadLang == "en" {
			fmt.Print("Fetching results: ")
		} else {
			fmt.Print("正在获取结果: ")
		}
		currentPage := 1
		fmt.Printf("%d ", currentPage)

		for len(allExploits) < limitedResults && pagination.HasMore() {
			// 获取下一页
			nextPage, err := pagination.GetNextPage()
			if err != nil {
				if payloadLang == "en" {
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
			if payloadMaxResults > 0 && len(allExploits) >= payloadMaxResults {
				allExploits = allExploits[:payloadMaxResults]
				break
			}

			// 打印进度
			currentPage++
			fmt.Printf("%d ", currentPage)
		}

		if payloadLang == "en" {
			fmt.Println("\nFetching completed")
		} else {
			fmt.Println("\n获取完成")
		}

		// 创建输出目录
		if payloadOutputDir == "" {
			// 如果没有指定输出目录，使用默认目录
			payloadOutputDir = filepath.Join("payloads", sanitizeFilename(query))
		}

		err = os.MkdirAll(payloadOutputDir, 0755)
		if err != nil {
			if payloadLang == "en" {
				fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
			}
			os.Exit(1)
		}

		// 保存每个结果
		if payloadLang == "en" {
			fmt.Printf("Saving %d exploits to %s...\n", len(allExploits), payloadOutputDir)
		} else {
			fmt.Printf("正在保存 %d 个漏洞利用到 %s...\n", len(allExploits), payloadOutputDir)
		}
		savedCount := 0

		for _, exploit := range allExploits {
			// 如果需要增强获取漏洞利用详情，可以添加额外的API调用
			// 例如: exploit = client.GetExploitDetail(exploit.ID)

			fileName, fileContent, err := prepareExploitFile(exploit, payloadLang)
			if err != nil {
				if payloadLang == "en" {
					fmt.Fprintf(os.Stderr, "Failed to process %s: %v\n", fileName, err)
				} else {
					fmt.Fprintf(os.Stderr, "处理 %s 失败: %v\n", fileName, err)
				}
				continue
			}

			filePath := filepath.Join(payloadOutputDir, fileName)

			// 保存到文件
			err = os.WriteFile(filePath, []byte(fileContent), 0644)
			if err != nil {
				if payloadLang == "en" {
					fmt.Fprintf(os.Stderr, "Failed to save %s: %v\n", fileName, err)
				} else {
					fmt.Fprintf(os.Stderr, "保存 %s 失败: %v\n", fileName, err)
				}
				continue
			}
			savedCount++
		}

		if payloadLang == "en" {
			fmt.Printf("Completed! Successfully saved %d/%d exploits\n", savedCount, len(allExploits))
			fmt.Printf("Exploits saved to: %s\n", payloadOutputDir)
		} else {
			fmt.Printf("完成! 成功保存 %d/%d 个漏洞利用\n", savedCount, len(allExploits))
			fmt.Printf("漏洞利用已保存到: %s\n", payloadOutputDir)
		}
	},
}

func init() {
	payloadCmd.Flags().StringVarP(&payloadSearchType, "type", "t", "", "搜索类型 (cve, title, tag)")
	payloadCmd.Flags().StringVarP(&payloadSortType, "sort", "s", "score", "排序方式 (score, date)")
	payloadCmd.Flags().IntVarP(&payloadMaxResults, "max", "m", 0, "最大结果数量 (0表示不限制)")
	payloadCmd.Flags().StringVarP(&payloadOutputDir, "output", "o", "", "输出目录 (默认: ./payloads/查询词)")
	payloadCmd.Flags().StringVarP(&payloadNaming, "naming", "n", "id", "文件命名方式 (id, title, both)")
	payloadCmd.Flags().StringVarP(&payloadLang, "lang", "l", "cn", "输出语言 (cn: 中文, en: 英文)")
}

// 根据漏洞语言选择正确的文件扩展名
func getFileExtension(language string) string {
	language = strings.ToLower(language)
	switch language {
	case "python", "py":
		return ".py"
	case "javascript", "js":
		return ".js"
	case "java":
		return ".java"
	case "php":
		return ".php"
	case "ruby", "rb":
		return ".rb"
	case "perl", "pl":
		return ".pl"
	case "bash", "shell", "sh":
		return ".sh"
	case "c":
		return ".c"
	case "c++", "cpp":
		return ".cpp"
	case "c#", "csharp":
		return ".cs"
	case "go", "golang":
		return ".go"
	case "powershell", "ps1":
		return ".ps1"
	case "sql":
		return ".sql"
	case "html":
		return ".html"
	case "xml":
		return ".xml"
	case "rust":
		return ".rs"
	case "swift":
		return ".swift"
	case "kotlin":
		return ".kt"
	case "typescript", "ts":
		return ".ts"
	default:
		return ".txt"
	}
}

// 获取基于语言的注释符号
func getCommentSymbol(language string) string {
	language = strings.ToLower(language)
	switch language {
	case "python", "py", "ruby", "rb", "perl", "pl", "bash", "shell", "sh", "powershell", "ps1":
		return "#"
	case "javascript", "js", "java", "c", "c++", "cpp", "c#", "csharp", "go", "golang", "php", "swift", "kotlin", "typescript", "ts", "rust":
		return "//"
	case "html", "xml":
		return "<!--"
	case "sql":
		return "--"
	default:
		return "#"
	}
}

// 获取基于语言的注释结束符号（如果有的话）
func getCommentEndSymbol(language string) string {
	language = strings.ToLower(language)
	if language == "html" || language == "xml" {
		return "-->"
	}
	return ""
}

// 为漏洞利用代码准备文件名和内容
func prepareExploitFile(exploit types.Exploit, outputLang string) (string, string, error) {
	timestamp := time.Now().Format("20060102_150405")
	var baseName string

	// 根据命名选项确定基本文件名
	switch payloadNaming {
	case "title":
		baseName = sanitizeFilename(exploit.Title)
	case "both":
		if exploit.ID != "" {
			baseName = sanitizeFilename(exploit.ID + "_" + exploit.Title)
		} else {
			baseName = sanitizeFilename(exploit.Title)
		}
	case "id", "":
		fallthrough
	default:
		if exploit.ID != "" {
			baseName = sanitizeFilename(exploit.ID)
		} else {
			baseName = sanitizeFilename(exploit.Title)
		}
	}

	// 根据漏洞语言选择正确的文件扩展名
	extension := getFileExtension(exploit.Language)

	// 完整的文件名
	fileName := fmt.Sprintf("%s_%s%s", baseName, timestamp, extension)

	// 获取此语言的注释符号
	commentSymbol := getCommentSymbol(exploit.Language)
	commentEndSymbol := getCommentEndSymbol(exploit.Language)

	// 创建文件内容，首先是元数据注释
	var contentBuilder strings.Builder

	// 添加元数据作为注释
	if commentEndSymbol != "" { // 对于HTML等需要结束标记的语言
		contentBuilder.WriteString(commentSymbol + "\n")
	}

	// 生成详情页URL（如果ID可用）
	var detailURL string
	if exploit.ID != "" {
		// 使用完整的ID，包括前缀
		detailURL = fmt.Sprintf("https://sploitus.com/exploit?id=%s", exploit.ID)
	}

	if outputLang == "en" {
		contentBuilder.WriteString(commentSymbol + " Title: " + exploit.Title + "\n")
		if exploit.ID != "" {
			// ID行已删除
			contentBuilder.WriteString(commentSymbol + " Sploitus Detail URL: " + detailURL + "\n")
		}
		if exploit.Score > 0 {
			contentBuilder.WriteString(commentSymbol + fmt.Sprintf(" Score: %.1f\n", exploit.Score))
		}
		if exploit.Published != "" {
			contentBuilder.WriteString(commentSymbol + " Published: " + exploit.Published + "\n")
		}
		if exploit.Href != "" {
			contentBuilder.WriteString(commentSymbol + " URL: " + exploit.Href + "\n")
		}
		if exploit.Type != "" {
			contentBuilder.WriteString(commentSymbol + " Type: " + exploit.Type + "\n")
		}
		if exploit.Language != "" {
			contentBuilder.WriteString(commentSymbol + " Language: " + exploit.Language + "\n")
		}

		// 添加额外的分隔行
		contentBuilder.WriteString(commentSymbol + " " + strings.Repeat("-", 50) + "\n")
		contentBuilder.WriteString(commentSymbol + " Exploit code below\n")
		contentBuilder.WriteString(commentSymbol + " " + strings.Repeat("-", 50) + "\n")
	} else {
		contentBuilder.WriteString(commentSymbol + " 标题 (Title): " + exploit.Title + "\n")
		if exploit.ID != "" {
			// ID行已删除
			contentBuilder.WriteString(commentSymbol + " Sploitus详情页 (Sploitus Detail URL): " + detailURL + "\n")
		}
		if exploit.Score > 0 {
			contentBuilder.WriteString(commentSymbol + fmt.Sprintf(" 得分 (Score): %.1f\n", exploit.Score))
		}
		if exploit.Published != "" {
			contentBuilder.WriteString(commentSymbol + " 发布日期 (Published): " + exploit.Published + "\n")
		}
		if exploit.Href != "" {
			contentBuilder.WriteString(commentSymbol + " 链接 (URL): " + exploit.Href + "\n")
		}
		if exploit.Type != "" {
			contentBuilder.WriteString(commentSymbol + " 类型 (Type): " + exploit.Type + "\n")
		}
		if exploit.Language != "" {
			contentBuilder.WriteString(commentSymbol + " 语言 (Language): " + exploit.Language + "\n")
		}

		// 添加额外的分隔行
		contentBuilder.WriteString(commentSymbol + " " + strings.Repeat("-", 50) + "\n")
		contentBuilder.WriteString(commentSymbol + " 以下是漏洞利用代码\n")
		contentBuilder.WriteString(commentSymbol + " " + strings.Repeat("-", 50) + "\n")
	}

	// 结束注释（如果需要）
	if commentEndSymbol != "" {
		contentBuilder.WriteString(commentEndSymbol + "\n\n")
	} else {
		contentBuilder.WriteString("\n")
	}

	// 添加原始源代码（如果有）
	if exploit.Source != "" {
		contentBuilder.WriteString(exploit.Source)
	} else {
		if outputLang == "en" {
			contentBuilder.WriteString(commentSymbol + " Note: No source code provided for this exploit\n")
			contentBuilder.WriteString(commentSymbol + " Please visit the URL above for more information\n")
		} else {
			contentBuilder.WriteString(commentSymbol + " 注意：此漏洞没有提供源代码内容\n")
			contentBuilder.WriteString(commentSymbol + " 请访问以上URL获取更多信息\n")
		}
	}

	return fileName, contentBuilder.String(), nil
}

// 清理文件名，替换非法字符
func sanitizeFilename(input string) string {
	// 替换非法字符
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)

	// 替换多个连续的下划线为单个下划线
	result := replacer.Replace(input)
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// 限制长度，避免文件名过长
	maxLength := 100
	if len(result) > maxLength {
		result = result[:maxLength]
	}

	return result
}
