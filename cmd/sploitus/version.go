package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 版本信息
var (
	Version   = "0.2.0"
	BuildDate = "unknown"
	GitHash   = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示程序版本信息",
	Long:  `显示Sploitus Crawler的版本、构建日期和Git哈希值`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sploitus Crawler")
		fmt.Printf("版本: %s\n", Version)
		fmt.Printf("构建日期: %s\n", BuildDate)
		fmt.Printf("Git Hash: %s\n", GitHash)
		fmt.Println("\n项目地址: https://github.com/scagogogo/sploitus-skills")
	},
}
