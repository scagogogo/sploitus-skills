# Sploitus Crawler

一个用于搜索和导出 [Sploitus](https://sploitus.com) 数据库中漏洞利用数据的命令行工具和Go库。

[English Version](#english-version)

## 功能特性

- 通过关键词、CVE ID或其他条件搜索漏洞利用
- 将结果导出为JSON文件
- 可配置的搜索参数（类型、排序、偏移量）
- 支持HTTP代理（可选）
- 便捷的分页功能
- 可作为命令行工具或Go库使用

## 安装

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/scagogogo/sploitus-skills.git
cd sploitus-skills

# 构建二进制文件
go build -o sploitus ./cmd/sploitus
```

## 使用方法

### 命令行界面

```bash
# 基本搜索
./sploitus search "CVE-2023-1234"

# 分页搜索，获取第2页，每页10条结果
./sploitus search "CVE-2023-1234" --type cve --sort score --page 2 --size 10

# 使用 list 命令自动获取所有结果
./sploitus list "wordpress"

# 限制获取结果数量
./sploitus list "wordpress" --max 50 --output all_wp_exploits.json

# 显示版本信息
./sploitus version

# 使用代理搜索（可选）
./sploitus search "CVE-2023-1234" --proxy http://localhost:8080

# 搜索并保存漏洞利用代码到本地文件夹
./sploitus payload "log4j" --output=./exploits --naming=both
```

### 可用命令

- `search` - 在Sploitus上搜索漏洞利用（支持分页）
- `list` - 列出所有匹配的结果（自动分页获取）
- `payload` - 搜索并保存漏洞利用代码到本地文件夹
- `version` - 显示程序版本信息

### 搜索命令参数

- `--type, -t` - 搜索类型 (cve, title, tag)
- `--sort, -s` - 结果排序 (score, date) [默认: score]
- `--page, -g` - 页码 [默认: 1]
- `--size, -n` - 每页结果数量 [默认: 10]
- `--output, -o` - 输出文件路径
- `--proxy, -p` - HTTP代理URL（例如：http://localhost:8080）【可选参数】
- `--format, -F` - 输出格式 (default, json, jq) [默认: default]
- `--pretty` - 美化JSON输出

### 列表命令参数

- `--type, -t` - 搜索类型 (cve, title, tag)
- `--sort, -s` - 结果排序 (score, date) [默认: score]
- `--output, -o` - 输出文件路径
- `--max, -m` - 最大结果数量，0表示不限制 [默认: 0]
- `--proxy, -p` - HTTP代理URL（例如：http://localhost:8080）【可选参数】
- `--format, -F` - 输出格式 (default, json, jq) [默认: default]
- `--pretty` - 美化JSON输出

### Payload 命令参数

- `--type, -t` - 搜索类型 (cve, title, tag)
- `--sort, -s` - 结果排序 (score, date) [默认: score]
- `--max, -m` - 最大结果数量，0表示不限制 [默认: 0]
- `--output, -o` - 输出目录 [默认: ./payloads/查询词]
- `--naming, -n` - 文件命名方式 (id, title, both) [默认: id]
- `--proxy, -p` - HTTP代理URL（例如：http://localhost:8080）【可选参数】
- `--format, -F` - 输出格式 (default, json, jq) [默认: default]
- `--pretty` - 美化JSON输出

**特性：**

- 根据漏洞利用的编程语言自动选择适当的文件扩展名（如 `.py`, `.js`, `.java` 等）
- 默认使用 `.txt` 扩展名当语言未知时
- 将漏洞元数据（标题、ID、链接等）作为注释添加到文件头部
- 使用适当的注释格式，基于文件类型（例如，`#` 对于Python，`//` 对于JavaScript）
- 将漏洞利用源代码（如果有）作为文件的主要内容

## Go库使用示例

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/scagogogo/sploitus-skills/pkg/sploitus"
)

func main() {
	// 创建新客户端（默认不使用代理）
	client := sploitus.NewClient()
	
	// 执行搜索
	response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}
	
	// 打印结果
	fmt.Printf("找到 %d 个结果\n", response.ExploitsTotal)
	for i, exploit := range response.Exploits {
		fmt.Printf("%d. %s (评分: %.1f)\n", i+1, exploit.Title, exploit.Score)
	}
	
	// 导出为JSON
	outputPath := "results.json"
	if err := sploitus.ExportJSON(response, outputPath); err != nil {
		log.Fatalf("保存结果失败: %v", err)
	}
	fmt.Printf("结果已保存到 %s\n", outputPath)
}
```

## 使用HTTP代理（可选功能）

代理功能是完全可选的。默认情况下，客户端会直接连接到目标服务器而不使用任何代理。只有当您明确指定代理时，客户端才会使用代理连接。

您可以通过以下两种方式使用HTTP代理：

### 方式1：创建时设置代理

```go
// 创建带有代理的客户端
client, err := sploitus.NewClientWithProxy("http://localhost:8080")
if err != nil {
    log.Fatalf("创建带代理的客户端失败: %v", err)
}

// 正常使用客户端
response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
```

### 方式2：在现有客户端上设置代理

```go
// 创建普通客户端
client := sploitus.NewClient()

// 设置代理
err := client.SetProxy("http://localhost:8080")
if err != nil {
    log.Fatalf("设置代理失败: %v", err)
}

// 正常使用客户端
response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
```

**错误处理说明：**
- 不使用代理不会报错
- 只有在显式设置了无效的代理URL时才会返回错误
- 可以随时更改或移除代理设置

完整示例请参见 [examples/proxy/proxy_example.go](examples/proxy/proxy_example.go)。

## 分页功能

Sploitus API支持分页，我们提供了便捷的分页助手来简化分页操作：

### 基本分页示例

```go
// 创建客户端
client := sploitus.NewClient()

// 创建分页助手
paginator := client.NewPaginationHelper("CVE-2023", "exploits", "default")

// 设置每页显示的结果数量（默认为20）
paginator.SetPageSize(10)

// 获取第一页结果
firstPage, err := paginator.GetFirstPage()
if err != nil {
    log.Fatalf("获取第一页失败: %v", err)
}

// 显示当前页信息
fmt.Println(paginator.GetPageInfo()) // 例如: "第 1/5 页, 总共 42 条结果"

// 获取下一页
if paginator.HasMore() {
    nextPage, err := paginator.GetNextPage()
    // 处理下一页结果...
}

// 直接跳到特定页码（页码从1开始）
page3, err := paginator.GetPage(3)
// 处理第3页结果...

// 获取所有结果（一次性，谨慎使用）
allResults, err := paginator.GetAllResults()
```

### 英文版分页助手

还提供了英文版的分页助手，功能与中文版完全相同：

```go
// 创建英文版分页助手
enPaginator := client.NewEnPaginationHelper("CVE-2023", "exploits", "default")
// 使用方法与中文版相同...
```

完整示例请参见 [examples/pagination/pagination_example.go](examples/pagination/pagination_example.go)。

## 许可证

详细信息请参见 [LICENSE](LICENSE) 文件。

---

<a name="english-version"></a>
# Sploitus Crawler (English Version)

A command-line tool and Go library for searching and exporting exploit data from the [Sploitus](https://sploitus.com) database.

## Features

- Search for exploits by keyword, CVE ID, or other criteria
- Export results to JSON files
- Configurable search parameters (type, sort, offset)
- Support for HTTP proxies (optional)
- Convenient pagination features
- Can be used as a command-line tool or Go library

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/scagogogo/sploitus-skills.git
cd sploitus-skills

# Build the binary
go build -o sploitus ./cmd/sploitus
```

## Usage

### Command-line Interface

```bash
# Basic search
./sploitus search "CVE-2023-1234"

# Paginated search, get page 2 with 10 results per page
./sploitus search "CVE-2023-1234" --type cve --sort score --page 2 --size 10

# Get all results with the list command
./sploitus list "wordpress"

# Limit the number of results
./sploitus list "wordpress" --max 50 --output all_wp_exploits.json

# Display version information
./sploitus version

# Search using a proxy (optional)
./sploitus search "CVE-2023-1234" --proxy http://localhost:8080

# Search and save exploit code to local directory
./sploitus payload "log4j" --output=./exploits --naming=both
```

### Available Commands

- `search` - Search for exploits on Sploitus (with pagination)
- `list` - List all matching results (auto-paginated)
- `payload` - Search and save exploit code to local directory
- `version` - Display program version information

### Search Command Flags

- `--type, -t` - Search type (cve, title, tag)
- `--sort, -s` - Sort results (score, date) [default: score]
- `--page, -g` - Page number [default: 1]
- `--size, -n` - Results per page [default: 10]
- `--output, -o` - Output file path
- `--proxy, -p` - HTTP proxy URL (e.g., http://localhost:8080) [optional]
- `--format, -F` - Output format (default, json, jq) [default: default]
- `--pretty` - Pretty-print JSON output

### List Command Flags

- `--type, -t` - Search type (cve, title, tag)
- `--sort, -s` - Sort results (score, date) [default: score]
- `--output, -o` - Output file path
- `--max, -m` - Maximum number of results, 0 means unlimited [default: 0]
- `--proxy, -p` - HTTP proxy URL (e.g., http://localhost:8080) [optional]
- `--format, -F` - Output format (default, json, jq) [default: default]
- `--pretty` - Pretty-print JSON output

### Payload Command Flags

- `--type, -t` - Search type (cve, title, tag)
- `--sort, -s` - Sort results (score, date) [default: score]
- `--max, -m` - Maximum number of results, 0 means unlimited [default: 0]
- `--output, -o` - Output directory [default: ./payloads/query]
- `--naming, -n` - File naming method (id, title, both) [default: id]
- `--proxy, -p` - HTTP proxy URL (e.g., http://localhost:8080) [optional]
- `--format, -F` - Output format (default, json, jq) [default: default]
- `--pretty` - Pretty-print JSON output

**Features:**

- Automatically selects appropriate file extensions based on exploit language (e.g., `.py`, `.js`, `.java`, etc.)
- Uses `.txt` extension by default when language is unknown
- Adds exploit metadata (title, ID, URL, etc.) as comments at the beginning of files
- Uses appropriate comment syntax based on file type (e.g., `#` for Python, `//` for JavaScript)
- Includes exploit source code (if available) as the main file content

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
	
	// Print results
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

## Using HTTP Proxies (Optional Feature)

The proxy feature is completely optional. By default, the client will connect directly to the target server without using any proxy. Only when you explicitly specify a proxy will the client use a proxy connection.

You can use HTTP proxies with the library in two ways:

### Method 1: Create a client with a proxy

```go
// Create a client with a proxy
client, err := sploitus.NewClientWithProxy("http://localhost:8080")
if err != nil {
    log.Fatalf("Failed to create client with proxy: %v", err)
}

// Use the client as usual
response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
```

### Method 2: Set a proxy on an existing client

```go
// Create a regular client
client := sploitus.NewClient()

// Set a proxy
err := client.SetProxy("http://localhost:8080")
if err != nil {
    log.Fatalf("Failed to set proxy: %v", err)
}

// Use the client as usual
response, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
```

**Error Handling Notes:**
- Not using a proxy will not cause any errors
- An error is returned only when an invalid proxy URL is explicitly set
- You can change or remove proxy settings at any time

For a complete example, see [examples/proxy/proxy_example.go](examples/proxy/proxy_example.go).

## Pagination Features

The Sploitus API supports pagination, and we provide convenient pagination helpers to simplify pagination operations:

### Basic Pagination Example

```go
// Create client
client := sploitus.NewClient()

// Create pagination helper
paginator := client.NewEnPaginationHelper("CVE-2023", "exploits", "default")

// Set results per page (default is 20)
paginator.SetPageSize(10)

// Get first page
firstPage, err := paginator.GetFirstPage()
if err != nil {
    log.Fatalf("Failed to get first page: %v", err)
}

// Show current page info
fmt.Println(paginator.GetPageInfo()) // e.g., "Page 1 of 5, Total 42 items"

// Get next page
if paginator.HasMore() {
    nextPage, err := paginator.GetNextPage()
    // Process next page results...
}

// Jump to specific page (page numbers start from 1)
page3, err := paginator.GetPage(3)
// Process page 3 results...

// Get all results (at once, use with caution)
allResults, err := paginator.GetAllResults()
```

### Chinese Version Pagination Helper

A Chinese version of the pagination helper is also available with identical functionality:

```go
// Create Chinese version pagination helper
cnPaginator := client.NewPaginationHelper("CVE-2023", "exploits", "default")
// Use methods same as English version...
```

For a complete example, see [examples/pagination/pagination_example.go](examples/pagination/pagination_example.go).

## License

See the [LICENSE](LICENSE) file for details. 