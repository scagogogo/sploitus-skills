package sploitus

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

// TestE2E_SearchWithRealCookie 使用真实 Cookie 进行端到端测试
// 需要设置环境变量 SPLOITUS_COOKIE（从浏览器复制）
// 如果未设置环境变量，测试自动跳过
func TestE2E_SearchWithRealCookie(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量（从浏览器复制 Cookie）")
	}

	client := NewClient()
	err := client.SetCookies(cookieStr)
	if err != nil {
		t.Fatalf("SetCookies failed: %v", err)
	}

	// 搜索 CVE 编号
	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Error("Expected at least 1 result")
	}
	t.Logf("Found %d total exploits", resp.ExploitsTotal)

	// 验证返回的字段
	for _, exp := range resp.Exploits {
		if exp.Title == "" {
			t.Error("Exploit title should not be empty")
		}
		if exp.Score <= 0 {
			t.Errorf("Exploit score should be positive, got %f", exp.Score)
		}
		if exp.ID == "" {
			t.Error("Exploit ID should not be empty")
		}
	}
}

// TestE2E_SearchByKeyword 使用真实 Cookie 搜索关键词
func TestE2E_SearchByKeyword(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	// 搜索关键词
	keywords := []string{"wordpress", "apache", "linux", "kernel"}
	for _, kw := range keywords {
		resp, err := client.Search(kw, "exploits", "default", 0)
		if err != nil {
			t.Errorf("Search %q failed: %v", kw, err)
			continue
		}
		t.Logf("Keyword %q: %d results", kw, resp.ExploitsTotal)
		if resp.ExploitsTotal > 0 && len(resp.Exploits) == 0 {
			t.Errorf("Search %q: has total but no items", kw)
		}
	}
}

// TestE2E_SearchWithPagination 翻页端到端测试
func TestE2E_SearchWithPagination(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	p := client.NewPaginationHelper("CVE-2024", "exploits", "default")
	p.SetPageSize(10)

	// 第1页
	page1, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}
	if page1.ExploitsTotal == 0 {
		t.Skip("No results for CVE-2024, skipping pagination test")
	}
	t.Logf("Page 1: %d items, total: %d", len(page1.Exploits), page1.ExploitsTotal)

	// 第2页
	if p.HasMore() {
		page2, err := p.GetNextPage()
		if err != nil {
			t.Fatalf("GetNextPage failed: %v", err)
		}
		t.Logf("Page 2: %d items", len(page2.Exploits))
	}
}

// TestE2E_GetExploitDetail 获取漏洞详情端到端测试
func TestE2E_GetExploitDetail(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	// 先搜索一个已知的漏洞，获取其 ID
	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Skip("No results for CVE-2025-1316")
	}

	// 获取详情
	exploitID := resp.Exploits[0].ID
	detail, err := client.GetExploitDetail(exploitID)
	if err != nil {
		t.Fatalf("GetExploitDetail(%s) failed: %v", exploitID, err)
	}
	if detail == nil {
		t.Fatal("Detail should not be nil")
	}
	t.Logf("Detail: ID=%s, Title=%s, Score=%.1f", detail.ID, detail.Title, detail.Score)

	if detail.Source == "" {
		t.Log("Detail source is empty (expected if no POC available)")
	}
}

// TestE2E_ExportJSON 导出端到端测试
func TestE2E_ExportJSON(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	resp, err := client.Search("test", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Skip("No results for 'test'")
	}

	// 导出
	outputPath := "/tmp/sploitus_e2e_export.json"
	err = ExportJSON(resp, outputPath)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// 验证导出内容
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	var loaded types.SearchResponse
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Unmarshal exported JSON failed: %v", err)
	}
	if loaded.ExploitsTotal != resp.ExploitsTotal {
		t.Errorf("Export total mismatch: %d vs %d", loaded.ExploitsTotal, resp.ExploitsTotal)
	}
	os.Remove(outputPath)
	t.Log("Export JSON verified successfully")
}

// TestE2E_ProxySearch 代理端到端测试
func TestE2E_ProxySearch(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	proxyURL := os.Getenv("SPLOITUS_PROXY")
	if proxyURL == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_PROXY 环境变量来测试代理")
	}

	client, err := NewClientWithProxy(proxyURL)
	if err != nil {
		t.Fatalf("NewClientWithProxy failed: %v", err)
	}
	client.SetCookies(cookieStr)

	resp, err := client.Search("CVE-2023-1234", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Proxy search failed: %v", err)
	}
	t.Logf("Proxy search: %d results", resp.ExploitsTotal)
}

// TestE2E_SearchWithAllTypes 测试所有搜索类型
func TestE2E_SearchWithAllTypes(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	searchTypes := []string{"exploits", "tools", ""}
	sorts := []string{"default", "score", "date"}

	for _, st := range searchTypes {
		for _, sort := range sorts {
			resp, err := client.Search("test", st, sort, 0)
			if err != nil {
				t.Errorf("Search(type=%q, sort=%q) failed: %v", st, sort, err)
				continue
			}
			t.Logf("type=%q sort=%q → %d total", st, sort, resp.ExploitsTotal)
		}
	}
}

// TestE2E_SearchPayloadCommand 验证 payload 命令能获取源码
func TestE2E_SearchPayloadCommand(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	// 搜索有 POC 的漏洞
	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Skip("No results")
	}

	// 验证 source 字段包含 POC 内容
	pocFound := false
	for _, exp := range resp.Exploits {
		if strings.Contains(exp.Source, "git clone") || strings.Contains(exp.Source, "```") {
			pocFound = true
			t.Logf("POC found: %s (source length: %d bytes)", exp.Title, len(exp.Source))
			break
		}
	}
	if !pocFound {
		t.Log("No POC source found in first page results")
	}
}

// TestE2E_GetExploitDetailBySearch 验证通过搜索获取详情
func TestE2E_GetExploitDetailBySearch(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	// 搜索结果应该包含所有需要的信息
	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Skip("No results")
	}

	for _, exp := range resp.Exploits {
		if exp.ID == "" {
			t.Error("Empty exploit ID")
		}
		if exp.Title == "" {
			t.Error("Empty exploit title")
		}
		if exp.Score <= 0 {
			t.Errorf("Non-positive score: %f", exp.Score)
		}
		if exp.Language == "" {
			t.Log("Empty language, likely not a direct code exploit")
		}
	}
}

// TestE2E_ListAllResults 自动翻页获取所有结果
func TestE2E_ListAllResults(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	all, err := p.GetAllResults()
	if err != nil {
		t.Fatalf("GetAllResults failed: %v", err)
	}
	t.Logf("GetAllResults: %d total items", len(all))

	if len(all) > 0 {
		t.Logf("First: %s, Last: %s", all[0].Title, all[len(all)-1].Title)
	}
}

// TestE2E_SearchByCVEOnly 纯 CVE 编号搜索
func TestE2E_SearchByCVEOnly(t *testing.T) {
	cookieStr := os.Getenv("SPLOITUS_COOKIE")
	if cookieStr == "" {
		t.Skip("SKIP: 需要设置 SPLOITUS_COOKIE 环境变量")
	}

	client := NewClient()
	client.SetCookies(cookieStr)

	cves := []string{"CVE-2025-1316", "CVE-2024-0001", "CVE-2023-1234"}
	for _, cve := range cves {
		resp, err := client.Search(cve, "exploits", "default", 0)
		if err != nil {
			t.Logf("CVE %s search: %v (may not exist)", cve, err)
			continue
		}
		t.Logf("CVE %s: %d results", cve, resp.ExploitsTotal)
	}
}
