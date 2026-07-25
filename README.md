# Sploitus Skills

[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/scagogogo/sploitus-skills)](https://goreportcard.com/report/github.com/scagogogo/sploitus-skills)

A command-line tool and Go library for searching, exporting, and downloading exploit data from the [Sploitus](https://sploitus.com) exploit database. Supports HTTP proxies, browser automation for CloudFlare bypass, pagination, and automatic payload saving with language-aware file extensions.

> **English** | [简体中文](#简体中文版)

## Features

- 🔍 Search exploits by keyword, CVE ID, or title
- 📄 Export results to JSON files
- 📑 Auto-paginated listing with `list` command
- 💾 Download exploit source code with `payload` command (auto-detects language extensions)
- 🌐 Browser automation mode to bypass CloudFlare protection
- 🔌 HTTP/HTTPS proxy support (optional)
- 📚 Pagination helper with both Chinese and English interfaces
- 📦 Usable as a Go library or CLI tool

## Installation

### From Source

```bash
git clone https://github.com/scagogogo/sploitus-skills.git
cd sploitus-skills
go build -o sploitus ./cmd/sploitus
```

## Quick Start

```bash
# Search for exploits
./sploitus search "CVE-2023-1234"

# Get all results (auto-paginated)
./sploitus list "wordpress" --max 50 --output results.json

# Download exploit source code
./sploitus payload "log4j" --output=./exploits

# Use browser automation for CloudFlare bypass
./sploitus search "CVE-2023-1234" --browser

# Use a proxy
./sploitus search "CVE-2023-1234" --proxy http://localhost:8080
```

## Commands

### `search` — Search exploits with pagination

```bash
./sploitus search [query] [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Search type (`cve`, `title`, `tag`) | |
| `--sort` | `-s` | Sort order (`score`, `date`) | `score` |
| `--page` | `-g` | Page number | `1` |
| `--size` | `-n` | Results per page | `10` |
| `--output` | `-o` | Output file path | |
| `--format` | `-F` | Output format (`default`, `json`, `jq`) | `default` |
| `--pretty` | | Pretty-print JSON | |
| `--proxy` | `-p` | HTTP proxy URL | |
| `--browser` | `-b` | Use browser automation | |
| `--debug-browser` | `-d` | Show browser window for debugging | |
| `--cookies` | | Authentication cookies | |
| `--lang` | | Output language (`cn`, `en`) | `cn` |

### `list` — Auto-paginated listing of all results

```bash
./sploitus list [query] [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Search type | |
| `--sort` | `-s` | Sort order | `score` |
| `--output` | `-o` | Output file path | |
| `--max` | `-m` | Maximum results (0 = unlimited) | `0` |
| `--browser` | `-b` | Use browser automation | |
| `--proxy` | `-p` | HTTP proxy URL | |

### `payload` — Search and download exploit source code

```bash
./sploitus payload [query] [flags]
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--type` | `-t` | Search type | |
| `--sort` | `-s` | Sort order | `score` |
| `--max` | `-m` | Maximum results | `0` |
| `--output` | `-o` | Output directory | `./payloads/<query>` |
| `--naming` | `-n` | File naming (`id`, `title`, `both`) | `id` |
| `--lang` | `-l` | Comment language | `cn` |
| `--browser` | `-b` | Use browser automation | |
| `--proxy` | `-p` | HTTP proxy URL | |

**Payload Features:**
- Automatically selects file extensions based on exploit language (`.py`, `.js`, `.java`, `.go`, `.rb`, `.sh`, `.php`, `.rs`, `.ts`, etc.)
- Falls back to `.txt` for unknown languages
- Adds exploit metadata (title, ID, score, URL) as file header comments
- Uses language-appropriate comment syntax (`#` for Python, `//` for Go/JS, `<!--` for HTML)
- Includes exploit source code when available

### `version` — Display version information

```bash
./sploitus version
```

## Go Library Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
)

func main() {
	// Create a new client (no proxy by default)
	client := sploitus.NewClient()

	// Perform a search
	response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Found %d results\n", response.ExploitsTotal)
	for i, exploit := range response.Exploits {
		fmt.Printf("%d. %s (Score: %.1f)\n", i+1, exploit.Title, exploit.Score)
	}

	// Export to JSON
	outputPath := "results.json"
	if err := sploitus.ExportJSON(response, outputPath); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}
	fmt.Printf("Results saved to %s\n", outputPath)
}
```

### Pagination

```go
client := sploitus.NewClient()

// Chinese pagination helper
paginator := client.NewPaginationHelper("CVE-2023", "exploits", "default")
paginator.SetPageSize(10)

// Get first page
firstPage, err := paginator.GetFirstPage()
if err != nil {
	log.Fatalf("GetFirstPage failed: %v", err)
}

// Iterate pages
for paginator.HasMore() {
	nextPage, err := paginator.GetNextPage()
	// Process results...
}

// Get all results at once
allResults, err := paginator.GetAllResults()

// English pagination helper
enPaginator := client.NewEnPaginationHelper("CVE-2023", "exploits", "default")
```

### HTTP Proxy

```go
// Method 1: Create client with proxy
client, err := sploitus.NewClientWithProxy("http://localhost:8080")

// Method 2: Set proxy on existing client
client := sploitus.NewClient()
err := client.SetProxy("http://localhost:8080")
```

### Browser Automation

```go
browser, err := sploitus.NewBrowserSearcher(false) // false = headless mode
defer browser.Close()
results, err := browser.Search("CVE-2023-1234", "exploits", "default", 0)
```

### Get Exploit Detail

```go
detail, err := client.GetExploitDetail("0147E6AA-6963-51CE-90F9-420346FA917B")
if err != nil {
	log.Fatalf("Failed to get exploit detail: %v", err)
}
fmt.Printf("Title: %s, Score: %.1f\n", detail.Title, detail.Score)
```

## Examples

Complete runnable examples are available in the [examples](examples/) directory:

- [Simple Search](examples/simple_search.go) — Basic search and JSON export
- [Proxy Usage](examples/proxy/proxy_example.go) — HTTP proxy examples
- [Pagination](examples/pagination/pagination_example.go) — Full pagination workflows

## Test Coverage

The project aims for **100% test coverage** of non-browser code. Browser automation tests require a real browser environment and are excluded from automatic coverage runs.

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## License

[MIT License](LICENSE)

---

<a name="简体中文版"></a>

# Sploitus Skills (简体中文版)

[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

一个用于搜索、导出和下载 [Sploitus](https://sploitus.com) 漏洞利用数据库数据的命令行工具和 Go 库。支持 HTTP 代理、浏览器自动化绕过 CloudFlare 防护、分页功能，以及按编程语言自动识别扩展名保存漏洞利用代码。

## 功能特性

- 🔍 按关键词、CVE ID 或标题搜索漏洞利用
- 📄 导出结果为 JSON 文件
- 📑 `list` 命令自动翻页获取所有结果
- 💾 `payload` 命令下载漏洞利用源代码（自动识别语言扩展名）
- 🌐 浏览器自动化模式绕过 CloudFlare 防护
- 🔌 可选 HTTP/HTTPS 代理支持
- 📚 中英文双语分页助手
- 📦 可作为 Go 库或 CLI 工具使用

## 安装

```bash
git clone https://github.com/scagogogo/sploitus-skills.git
cd sploitus-skills
go build -o sploitus ./cmd/sploitus
```

## 快速开始

```bash
# 搜索漏洞利用
./sploitus search "CVE-2023-1234"

# 获取所有结果（自动翻页）
./sploitus list "wordpress" --max 50 --output results.json

# 下载漏洞利用源代码
./sploitus payload "log4j" --output=./exploits

# 使用浏览器绕过 CloudFlare
./sploitus search "CVE-2023-1234" --browser

# 使用代理
./sploitus search "CVE-2023-1234" --proxy http://localhost:8080
```

## 命令

### `search` — 分页搜索漏洞利用

```bash
./sploitus search [查询词] [参数]
```

| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--type` | `-t` | 搜索类型 (`cve`, `title`, `tag`) | |
| `--sort` | `-s` | 排序方式 (`score`, `date`) | `score` |
| `--page` | `-g` | 页码 | `1` |
| `--size` | `-n` | 每页结果数 | `10` |
| `--output` | `-o` | 输出文件路径 | |
| `--format` | `-F` | 输出格式 (`default`, `json`, `jq`) | `default` |
| `--pretty` | | 美化 JSON 输出 | |
| `--proxy` | `-p` | HTTP 代理 URL | |
| `--browser` | `-b` | 使用浏览器自动化 | |
| `--debug-browser` | `-d` | 显示浏览器窗口（调试用） | |
| `--cookies` | | 认证 Cookie | |
| `--lang` | | 输出语言 (`cn`, `en`) | `cn` |

### `list` — 自动翻页获取所有结果

```bash
./sploitus list [查询词] [参数]
```

| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--type` | `-t` | 搜索类型 | |
| `--sort` | `-s` | 排序方式 | `score` |
| `--output` | `-o` | 输出文件路径 | |
| `--max` | `-m` | 最大结果数（0=不限制） | `0` |
| `--browser` | `-b` | 使用浏览器自动化 | |
| `--proxy` | `-p` | HTTP 代理 URL | |

### `payload` — 搜索并保存漏洞利用代码

```bash
./sploitus payload [查询词] [参数]
```

| 参数 | 缩写 | 说明 | 默认值 |
|------|------|------|--------|
| `--type` | `-t` | 搜索类型 | |
| `--sort` | `-s` | 排序方式 | `score` |
| `--max` | `-m` | 最大结果数 | `0` |
| `--output` | `-o` | 输出目录 | `./payloads/<查询词>` |
| `--naming` | `-n` | 文件命名方式 (`id`, `title`, `both`) | `id` |
| `--lang` | `-l` | 注释语言 | `cn` |
| `--browser` | `-b` | 使用浏览器自动化 | |
| `--proxy` | `-p` | HTTP 代理 URL | |

**Payload 特性：**
- 根据漏洞利用编程语言自动选择扩展名（`.py`, `.js`, `.java`, `.go`, `.rb`, `.sh`, `.php`, `.rs`, `.ts` 等）
- 未知语言默认使用 `.txt` 扩展名
- 在文件头部添加漏洞元数据（标题、ID、得分、URL）作为注释
- 根据语言类型使用合适的注释符号（Python 用 `#`，Go/JS 用 `//`，HTML 用 `<!--`）
- 包含漏洞利用源代码（如果有）

### `version` — 显示版本信息

```bash
./sploitus version
```

## Go 库使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
)

func main() {
	// 创建新客户端（默认无代理）
	client := sploitus.NewClient()

	// 执行搜索
	response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}

	fmt.Printf("找到 %d 个结果\n", response.ExploitsTotal)
	for i, exploit := range response.Exploits {
		fmt.Printf("%d. %s (评分: %.1f)\n", i+1, exploit.Title, exploit.Score)
	}

	// 导出为 JSON
	outputPath := "results.json"
	if err := sploitus.ExportJSON(response, outputPath); err != nil {
		log.Fatalf("保存结果失败: %v", err)
	}
	fmt.Printf("结果已保存到 %s\n", outputPath)
}
```

### 分页功能

```go
client := sploitus.NewClient()

// 中文分页助手
paginator := client.NewPaginationHelper("CVE-2023", "exploits", "default")
paginator.SetPageSize(10)

// 获取第一页
firstPage, err := paginator.GetFirstPage()
if err != nil {
	log.Fatalf("获取第一页失败: %v", err)
}

// 遍历所有页
for paginator.HasMore() {
	nextPage, err := paginator.GetNextPage()
	// 处理结果...
}

// 一次性获取所有结果
allResults, err := paginator.GetAllResults()

// 英文分页助手
enPaginator := client.NewEnPaginationHelper("CVE-2023", "exploits", "default")
```

### HTTP 代理

```go
// 方式1：创建时设置代理
client, err := sploitus.NewClientWithProxy("http://localhost:8080")

// 方式2：在现有客户端上设置代理
client := sploitus.NewClient()
err := client.SetProxy("http://localhost:8080")
```

### 浏览器自动化

```go
browser, err := sploitus.NewBrowserSearcher(false) // false = 无头模式
defer browser.Close()
results, err := browser.Search("CVE-2023-1234", "exploits", "default", 0)
```

### 获取漏洞详情

```go
detail, err := client.GetExploitDetail("0147E6AA-6963-51CE-90F9-420346FA917B")
if err != nil {
	log.Fatalf("获取详情失败: %v", err)
}
fmt.Printf("标题: %s, 得分: %.1f\n", detail.Title, detail.Score)
```

## 示例

可运行的完整示例在 [examples](examples/) 目录中：

- [简单搜索](examples/simple_search.go) — 基础搜索和 JSON 导出
- [代理使用](examples/proxy/proxy_example.go) — HTTP 代理示例
- [分页](examples/pagination/pagination_example.go) — 完整分页工作流

## 测试覆盖

项目目标是对非浏览器代码实现 **100% 测试覆盖**。浏览器自动化测试需要真实浏览器环境，在自动覆盖率统计中排除。

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 许可证

[MIT License](LICENSE)