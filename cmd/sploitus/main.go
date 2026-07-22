package main

import (
	"fmt"
	"os"

	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
	"github.com/spf13/cobra"
)

var (
	// 全局标志
	proxyURL     string
	outputFormat string
	prettyPrint  bool
	language     string // 全局语言设置
	cookiesStr   string // 认证Cookies
	useBrowser   bool   // 是否使用浏览器自动化
	debugBrowser bool   // 是否启用浏览器调试模式
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sploitus",
	Short: "Sploitus Crawler - 搜索和导出Sploitus漏洞数据库",
	Long: `Sploitus Crawler 是一个命令行工具，用于搜索和导出Sploitus数据库中的漏洞利用数据。
它提供了命令行界面，同时也可以作为Go库使用。

可用命令:
  search      搜索漏洞利用（支持分页）
  list        列出搜索结果（自动翻页获取所有结果）
  payload     搜索并保存漏洞利用代码到本地文件夹
  version     显示版本信息`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供子命令，显示帮助
		cmd.Help()
	},
}

func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringVarP(&proxyURL, "proxy", "p", "", "HTTP代理URL (例如: http://localhost:8080)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "F", "default", "输出格式 (default, json, jq)")
	rootCmd.PersistentFlags().BoolVar(&prettyPrint, "pretty", false, "美化JSON输出")
	rootCmd.PersistentFlags().StringVar(&language, "lang", "cn", "输出语言 (cn: 中文, en: 英文)")
	rootCmd.PersistentFlags().StringVar(&cookiesStr, "cookies", "", "认证Cookies字符串 (从浏览器中复制)")
	rootCmd.PersistentFlags().BoolVar(&useBrowser, "browser", false, "使用自动化浏览器绕过CloudFlare保护")
	rootCmd.PersistentFlags().BoolVar(&debugBrowser, "debug-browser", false, "启用浏览器调试模式（显示浏览器窗口）")

	// 添加子命令
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(payloadCmd)
	rootCmd.AddCommand(versionCmd)
}

// 创建新客户端，考虑代理设置
func createClient() (*sploitus.Client, error) {
	var client *sploitus.Client
	var err error

	if proxyURL != "" {
		client, err = sploitus.NewClientWithProxy(proxyURL)
		if err != nil {
			if language == "en" {
				return nil, fmt.Errorf("failed to create client with proxy: %w", err)
			}
			return nil, fmt.Errorf("创建代理客户端失败: %w", err)
		}
	} else {
		client = sploitus.NewClient()
	}

	// 设置cookies
	if cookiesStr != "" {
		if err := client.SetCookies(cookiesStr); err != nil {
			if language == "en" {
				return nil, fmt.Errorf("failed to set cookies: %w", err)
			}
			return nil, fmt.Errorf("设置cookies失败: %w", err)
		}
	}

	return client, nil
}
