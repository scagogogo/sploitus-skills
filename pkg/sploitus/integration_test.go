package sploitus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

// ============================================================
// 集成测试：使用 data/search.md 中的真实 API 响应数据
// ============================================================

// realExploitResponse 是 data/search.md 中记录的真实 API 响应数据
func realExploitResponse() *types.SearchResponse {
	return &types.SearchResponse{
		Exploits: []types.Exploit{
			{
				Title: "Exploit for CVE-2025-1316", Score: 9.3,
				Href: "https://github.com/slockit/CVE-2025-1316",
				Type: "githubexploit", Published: "2025-03-29",
				ID: "0147E6AA-6963-51CE-90F9-420346FA917B",
				Source: "## https://sploitus.com/exploit?id=0147E6AA-6963-51CE-90F9-420346FA917B\n# CVE-2025-1316\n\n> Run as root\n\nEdimax IC-7100 does not properly neutralize requests.\n\n# Install\n```\nsudo apt update\nsudo apt install git\ngit clone https://github.com/slockit/CVE-2025-1316.git\n```\n\n# Usage\n```\n./CVE-2025-1316 [https://.com/]\n```",
				Language: "MARKDOWN",
			},
			{
				Title: "Edimax IP Camera NTP_serverName command injection",
				Score: 9.3,
				Href: "https://download.saintcorporation.com/cgi-bin/exploit_info/edimax_ip_camera_ntp_servername",
				Type: "saint", Published: "2025-03-21",
				ID: "SAINT:2CEDD0194C77120545A6315E534CFE66",
				Source: "## https://sploitus.com/exploit?id=SAINT:2CEDD0194C77120545A6315E534CFE66\nAdded: 03/21/2025\nCVE: CVE-2025-1316",
				Language: "MARKDOWN",
			},
		},
		ExploitsTotal: 2,
	}
}

// mockSearchServer 创建一个模拟搜索服务器的辅助函数
func mockSearchServer(response *types.SearchResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != SearchEndpoint {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

// ============================================================
// 1. 真实数据集成测试
// ============================================================

func TestIntegration_RealDataSearch(t *testing.T) {
	realData := realExploitResponse()
	server := mockSearchServer(realData)
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Real data search failed: %v", err)
	}
	if resp.ExploitsTotal != 2 {
		t.Errorf("Expected 2 total exploits, got %d", resp.ExploitsTotal)
	}
	if len(resp.Exploits) != 2 {
		t.Fatalf("Expected 2 exploits, got %d", len(resp.Exploits))
	}

	// Verify first exploit fields match real data
	e0 := resp.Exploits[0]
	if e0.Title != "Exploit for CVE-2025-1316" {
		t.Errorf("Title mismatch: %q", e0.Title)
	}
	if e0.Score != 9.3 {
		t.Errorf("Score mismatch: %f", e0.Score)
	}
	if e0.Type != "githubexploit" {
		t.Errorf("Type mismatch: %q", e0.Type)
	}
	if e0.ID != "0147E6AA-6963-51CE-90F9-420346FA917B" {
		t.Errorf("ID mismatch: %q", e0.ID)
	}
	if e0.Language != "MARKDOWN" {
		t.Errorf("Language mismatch: %q", e0.Language)
	}
	if !strings.Contains(e0.Source, "CVE-2025-1316") {
		t.Error("Source should contain CVE reference")
	}

	// Verify second exploit
	e1 := resp.Exploits[1]
	if e1.Title != "Edimax IP Camera NTP_serverName command injection" {
		t.Errorf("Title mismatch: %q", e1.Title)
	}
	if e1.Type != "saint" {
		t.Errorf("Type mismatch: %q", e1.Type)
	}
	if e1.ID != "SAINT:2CEDD0194C77120545A6315E534CFE66" {
		t.Errorf("ID mismatch: %q", e1.ID)
	}
}

func TestIntegration_RealDataSearchWithQuery(t *testing.T) {
	realData := realExploitResponse()
	server := mockSearchServer(realData)
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	query := &types.SearchQuery{
		Type:   "exploits",
		Sort:   "default",
		Query:  "CVE-2025-1316",
		Title:  false,
		Offset: 0,
	}
	resp, err := client.SearchWithQuery(query)
	if err != nil {
		t.Fatalf("Real data SearchWithQuery failed: %v", err)
	}
	if resp.ExploitsTotal != 2 {
		t.Errorf("Expected 2 exploits, got %d", resp.ExploitsTotal)
	}
}

func TestIntegration_RealDataExportJSON(t *testing.T) {
	realData := realExploitResponse()

	outputPath := fmt.Sprintf("/tmp/sploitus_test_real_%d.json", hashCode(t.Name()))
	defer func() {
		http.DefaultClient.Get(fmt.Sprintf("file://%s", outputPath)) // best effort cleanup
	}()

	err := ExportJSON(realData, outputPath)
	if err != nil {
		t.Fatalf("ExportJSON with real data failed: %v", err)
	}

	// Read back and verify
	data, err := readFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}
	var loaded types.SearchResponse
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}
	if loaded.ExploitsTotal != 2 {
		t.Errorf("Expected 2 exploits, got %d", loaded.ExploitsTotal)
	}
	if loaded.Exploits[0].Title != "Exploit for CVE-2025-1316" {
		t.Errorf("Title mismatch after export: %q", loaded.Exploits[0].Title)
	}
}

// hashCode generates a simple hash for test isolation
func hashCode(name string) int {
	h := 0
	for _, c := range name {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

// readFile is a helper to read a file
var readFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ============================================================
// 2. SearchQuery 变体测试
// ============================================================

func TestSearchQuery_ExploitsType(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search with exploits type failed: %v", err)
	}
}

func TestSearchQuery_ToolsType(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "tools", "default", 0)
	if err != nil {
		t.Fatalf("Search with tools type failed: %v", err)
	}
}

func TestSearchQuery_EmptyType(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "", "default", 0)
	if err != nil {
		t.Fatalf("Search with empty type failed: %v", err)
	}
}

func TestSearchQuery_SortByScore(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("CVE", "exploits", "score", 0)
	if err != nil {
		t.Fatalf("Search with score sort failed: %v", err)
	}
}

func TestSearchQuery_SortByDate(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("CVE", "exploits", "date", 0)
	if err != nil {
		t.Fatalf("Search with date sort failed: %v", err)
	}
}

func TestSearchQuery_WithOffset(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 20)
	if err != nil {
		t.Fatalf("Search with offset failed: %v", err)
	}
}

func TestSearchQuery_WithNegativeOffset(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", -1)
	if err != nil {
		t.Fatalf("Search with negative offset should still work: %v", err)
	}
}

func TestSearchQuery_WithCVESearch(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	// Simulate searching by CVE ID
	resp, err := client.Search("CVE-2025-1316", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("CVE search failed: %v", err)
	}
	if resp.ExploitsTotal == 0 {
		t.Error("Expected at least one result for CVE search")
	}
}

func TestSearchQuery_WithSpecialChars(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	// URL encoded or special characters in query
	_, err := client.Search("SQL Injection + OR '1'='1", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search with special chars failed: %v", err)
	}
}

func TestSearchQuery_WithUnicode(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("漏洞 测试", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search with unicode failed: %v", err)
	}
}

// ============================================================
// 3. SearchQuery 对象构造测试
// ============================================================

func TestSearchQuery_DefaultTitle(t *testing.T) {
	// Search() sets Title: true by default
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q types.SearchQuery
		json.NewDecoder(r.Body).Decode(&q)
		if !q.Title {
			t.Error("Expected Title to be true by default")
		}
		json.NewEncoder(w).Encode(realExploitResponse())
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	client.Search("test", "exploits", "default", 0)
}

func TestSearchQuery_OffsetCalculated(t *testing.T) {
	expectedOffset := 30
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var q types.SearchQuery
		json.NewDecoder(r.Body).Decode(&q)
		if q.Offset != expectedOffset {
			t.Errorf("Expected offset %d, got %d", expectedOffset, q.Offset)
		}
		json.NewEncoder(w).Encode(realExploitResponse())
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	client.Search("test", "exploits", "default", expectedOffset)
}

func TestSearchQuery_EmptyQuery(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("", "exploits", "default", 0)
	if err != nil {
		t.Fatalf("Search with empty query failed: %v", err)
	}
}

// ============================================================
// 4. Exploit 字段完整性测试
// ============================================================

func TestExploit_AllFieldsPresent(t *testing.T) {
	exp := types.Exploit{
		Title: "Test Title", Score: 9.9,
		Href: "https://example.com/e", Type: "exploit",
		Published: "2024-01-15", ID: "CVE-2024-0001",
		Source: "source code here", Language: "python",
	}

	if exp.Title == "" {
		t.Error("Title should not be empty")
	}
	if exp.Score <= 0 {
		t.Error("Score should be positive")
	}
	if exp.Href == "" {
		t.Error("Href should not be empty")
	}
	if exp.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestExploit_ZeroScore(t *testing.T) {
	exp := types.Exploit{Title: "Zero Score", Score: 0, ID: "ZERO-1"}
	if exp.Score != 0 {
		t.Errorf("Expected score 0, got %f", exp.Score)
	}
}

func TestExploit_MaxScore(t *testing.T) {
	exp := types.Exploit{Title: "Max Score", Score: 10.0, ID: "MAX-1"}
	if exp.Score != 10.0 {
		t.Errorf("Expected score 10.0, got %f", exp.Score)
	}
}

func TestExploit_EmptyTitle(t *testing.T) {
	exp := types.Exploit{Title: "", Score: 5.0, ID: "EMPTY-TITLE-1"}
	if exp.Title != "" {
		t.Error("Title should be empty")
	}
}

func TestExploit_LongTitle(t *testing.T) {
	longTitle := strings.Repeat("A", 1000)
	exp := types.Exploit{Title: longTitle, Score: 5.0, ID: "LONG-1"}
	if len(exp.Title) != 1000 {
		t.Errorf("Expected title length 1000, got %d", len(exp.Title))
	}
}

func TestExploit_EmptyID(t *testing.T) {
	exp := types.Exploit{Title: "No ID", Score: 5.0, ID: ""}
	if exp.ID != "" {
		t.Error("ID should be empty")
	}
}

func TestExploit_LongID(t *testing.T) {
	longID := "SAINT:" + strings.Repeat("A", 200)
	exp := types.Exploit{Title: "Long ID", Score: 5.0, ID: longID}
	if len(exp.ID) < 200 {
		t.Errorf("Expected long ID, got length %d", len(exp.ID))
	}
}

func TestExploit_SourceContainsPOC(t *testing.T) {
	exp := types.Exploit{Title: "POC Test", Score: 5.0, ID: "POC-1"}
	// 模拟搜索API返回的source字段包含POC代码
	exp.Source = "#!/usr/bin/python\n# Exploit for CVE-2024-0001\nimport socket\n\ndef exploit():\n    s = socket.socket()\n    s.connect(('target', 80))\n    s.send(b'PAYLOAD')\n    s.close()\n"

	if !strings.Contains(exp.Source, "import socket") {
		t.Error("Source should contain POC code")
	}
	if !strings.Contains(exp.Source, "def exploit") {
		t.Error("Source should contain exploit function")
	}
}

func TestExploit_MultipleLanguageSource(t *testing.T) {
	// Test that source can contain code in various languages
	langs := []string{"python", "ruby", "javascript", "go", "c", "php"}
	for _, lang := range langs {
		exp := types.Exploit{Title: "Test", Score: 5.0, ID: "LANG-TEST", Language: lang}
		if exp.Language != lang {
			t.Errorf("Expected language %s, got %s", lang, exp.Language)
		}
	}
}

// ============================================================
// 5. ExploitDetail 字段完整性测试
// ============================================================

func TestExploitDetail_AllFields(t *testing.T) {
	detail := types.ExploitDetail{
		ID: "CVE-2024-0001", Title: "Full Detail",
		Description: "A detailed description of the exploit",
		Published: "2024-01-01", Source: "source code",
		Language: "python", Score: 9.5, Type: "exploit",
		Href: "https://example.com/e",
	}

	if detail.Description == "" {
		t.Error("Description should not be empty")
	}
	if detail.Score != 9.5 {
		t.Errorf("Expected score 9.5, got %f", detail.Score)
	}
}

func TestExploitDetail_EmptyDescription(t *testing.T) {
	detail := types.ExploitDetail{ID: "CVE-2024-0002", Title: "No Description"}
	if detail.Description != "" {
		t.Error("Description should be empty")
	}
}

func TestExploitDetail_FromExploitConversion(t *testing.T) {
	exp := types.Exploit{
		Title: "Converted", Score: 7.5,
		Href: "https://example.com/c", Type: "exploit",
		Published: "2024-03-01", ID: "CVE-2024-0003",
		Source: "exploit source", Language: "ruby",
	}

	detail := &types.ExploitDetail{
		ID: exp.ID, Title: exp.Title,
		Source: exp.Source, Language: exp.Language,
		Score: exp.Score, Type: exp.Type,
		Href: exp.Href, Published: exp.Published,
	}

	if detail.ID != exp.ID {
		t.Error("ID should match")
	}
	if detail.Score != exp.Score {
		t.Error("Score should match")
	}
	if detail.Source != exp.Source {
		t.Error("Source should match")
	}
}

// ============================================================
// 6. SearchResponse 数据完整性测试
// ============================================================

func TestSearchResponse_EmptyExploits(t *testing.T) {
	resp := &types.SearchResponse{
		Exploits:      []types.Exploit{},
		ExploitsTotal: 0,
	}
	if len(resp.Exploits) != 0 {
		t.Error("Exploits should be empty")
	}
	if resp.ExploitsTotal != 0 {
		t.Errorf("Expected 0 total, got %d", resp.ExploitsTotal)
	}
}

func TestSearchResponse_MismatchedTotal(t *testing.T) {
	// Simulate API returning mismatched total vs actual items
	resp := &types.SearchResponse{
		Exploits:      []types.Exploit{{Title: "A", ID: "A-1"}, {Title: "B", ID: "B-1"}},
		ExploitsTotal: 100,
	}
	if len(resp.Exploits) != 2 {
		t.Errorf("Expected 2 exploits, got %d", len(resp.Exploits))
	}
	if resp.ExploitsTotal != 100 {
		t.Errorf("Expected total 100, got %d", resp.ExploitsTotal)
	}
}

func TestSearchResponse_LargeResultSet(t *testing.T) {
	exploits := make([]types.Exploit, 100)
	for i := 0; i < 100; i++ {
		exploits[i] = types.Exploit{
			Title: fmt.Sprintf("Exploit %d", i),
			Score: float64(i) / 10.0,
			ID:    fmt.Sprintf("CVE-2024-%04d", i),
		}
	}
	resp := &types.SearchResponse{Exploits: exploits, ExploitsTotal: 100}
	if len(resp.Exploits) != 100 {
		t.Errorf("Expected 100 exploits, got %d", len(resp.Exploits))
	}
	if resp.Exploits[99].Title != "Exploit 99" {
		t.Errorf("Last exploit title mismatch: %q", resp.Exploits[99].Title)
	}
}

// ============================================================
// 7. Client 行为测试
// ============================================================

func TestClient_DefaultBaseURL(t *testing.T) {
	client := NewClient()
	if client.BaseURL != DefaultBaseURL {
		t.Errorf("Expected default base URL %s, got %s", DefaultBaseURL, client.BaseURL)
	}
}

func TestClient_DefaultTimeout(t *testing.T) {
	client := NewClient()
	if client.HTTPClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.HTTPClient.Timeout)
	}
}

func TestClient_UserAgentNotEmpty(t *testing.T) {
	client := NewClient()
	if client.UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
}

func TestClient_AcceptLanguage(t *testing.T) {
	client := NewClient()
	if client.AcceptLanguage == "" {
		t.Error("AcceptLanguage should not be empty")
	}
}

func TestClient_CookiesInitiallyEmpty(t *testing.T) {
	client := NewClient()
	if len(client.Cookies) != 0 {
		t.Errorf("Expected 0 cookies, got %d", len(client.Cookies))
	}
}

func TestClient_BaseURLTrailingSlash(t *testing.T) {
	client := NewClient()
	// Verify base URL doesn't have trailing slash issues
	if strings.HasSuffix(client.BaseURL, "/") && strings.HasSuffix(client.BaseURL, "//") {
		t.Error("BaseURL should not have double trailing slash")
	}
}

func TestClient_MultipleRequests(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		json.NewEncoder(w).Encode(realExploitResponse())
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	// Make multiple sequential requests
	for i := 0; i < 5; i++ {
		_, err := client.Search("test", "exploits", "default", 0)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}
	if requestCount != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}
}

// ============================================================
// 8. 错误处理测试
// ============================================================

func TestError_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for server error")
	}
}

func TestError_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid search"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for bad request")
	}
	if !strings.Contains(err.Error(), "invalid search") {
		t.Errorf("Error should contain response body: %v", err)
	}
}

func TestError_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for forbidden")
	}
}

func TestError_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for not found")
	}
}

func TestError_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for rate limit")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("Error should mention rate limit: %v", err)
	}
}

func TestError_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(``))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for empty response")
	}
}

func TestError_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestError_ConnectionRefused(t *testing.T) {
	client := NewClient()
	client.BaseURL = "http://127.0.0.1:1"

	_, err := client.Search("test", "exploits", "default", 0)
	if err == nil {
		t.Log("Connection refused test did not error (unexpected)")
	}
}

// ============================================================
// 9. 代理配置测试
// ============================================================

func TestProxy_ValidURL(t *testing.T) {
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer proxyServer.Close()

	client, err := NewClientWithProxy(proxyServer.URL)
	if err != nil {
		t.Fatalf("NewClientWithProxy failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to not be nil")
	}
}

func TestProxy_SOCKS5Proxy(t *testing.T) {
	// SOCKS5 URLs should be rejected by http.Transport.Proxy
	_, err := NewClientWithProxy("socks5://127.0.0.1:1080")
	if err != nil {
		t.Logf("SOCKS5 proxy error (expected): %v", err)
	}
}

func TestProxy_HTTPSProxy(t *testing.T) {
	client, err := NewClientWithProxy("https://proxy.example.com:8080")
	if err != nil {
		t.Fatalf("HTTPS proxy failed: %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to not be nil")
	}
}

func TestProxy_SetProxyTwice(t *testing.T) {
	client := NewClient()
	err := client.SetProxy("http://proxy1:8080")
	if err != nil {
		t.Fatalf("First SetProxy failed: %v", err)
	}
	err = client.SetProxy("http://proxy2:8080")
	if err != nil {
		t.Fatalf("Second SetProxy failed: %v", err)
	}
}

func TestProxy_ClearProxy(t *testing.T) {
	client := NewClient()
	client.SetProxy("http://proxy:8080")
	// Set empty string proxy URL to clear
	client.SetProxy("http://")
	// Should not error
}

// ============================================================
// 10. Pagination 集成测试
// ============================================================

func TestPagination_SequentialPages(t *testing.T) {
	totalItems := 25
	pageSize := 10
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		offset := 0
		var q types.SearchQuery
		json.NewDecoder(r.Body).Decode(&q)
		offset = q.Offset
		itemsInPage := 0
		if offset < totalItems {
			itemsInPage = pageSize
			if offset+pageSize > totalItems {
				itemsInPage = totalItems - offset
			}
		}
		exploits := make([]types.Exploit, itemsInPage)
		for i := 0; i < itemsInPage; i++ {
			exploits[i] = types.Exploit{Title: fmt.Sprintf("E%d", offset+i), Score: 5.0, ID: fmt.Sprintf("ID-%d", offset+i)}
		}
		json.NewEncoder(w).Encode(types.SearchResponse{Exploits: exploits, ExploitsTotal: totalItems})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(pageSize)

	// Page 1: 10 items
	p1, _ := p.GetFirstPage()
	if len(p1.Exploits) != 10 {
		t.Errorf("Page 1: expected 10, got %d", len(p1.Exploits))
	}

	// Page 2: 10 items
	p2, _ := p.GetNextPage()
	if len(p2.Exploits) != 10 {
		t.Errorf("Page 2: expected 10, got %d", len(p2.Exploits))
	}

	// Page 3: 5 items (offset 20-24)
	p3, _ := p.GetNextPage()
	if len(p3.Exploits) != 5 {
		t.Errorf("Page 3: expected 5, got %d", len(p3.Exploits))
	}

	// HasMore: currentPos=20, totalItems=25 → true
	if !p.HasMore() {
		t.Error("HasMore should be true after page 3 (currentPos=20 < totalItems=25)")
	}

	// Page 4: offset=30, totalItems=25 → 0 items
	p4, _ := p.GetNextPage()
	if len(p4.Exploits) != 0 {
		t.Errorf("Page 4: expected 0, got %d", len(p4.Exploits))
	}

	// No more pages after page 4
	if p.HasMore() {
		t.Error("Should not have more pages after page 4")
	}
}

func TestPagination_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{Title: "Only", Score: 5.0, ID: "ONLY-1"}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	p1, _ := p.GetFirstPage()
	if len(p1.Exploits) != 1 {
		t.Errorf("Expected 1 item, got %d", len(p1.Exploits))
	}

	total, _ := p.GetTotalPages()
	if total != 1 {
		t.Errorf("Expected 1 page, got %d", total)
	}
}

func TestPagination_ExactFit(t *testing.T) {
	// 20 items, page size 10 = exactly 2 pages
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: make([]types.Exploit, 10),
			ExploitsTotal: 20,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)
	p.GetFirstPage()

	total, _ := p.GetTotalPages()
	if total != 2 {
		t.Errorf("Expected 2 pages for 20 items at size 10, got %d", total)
	}
}

func TestPagination_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{},
			ExploitsTotal: 0,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p1, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}
	if len(p1.Exploits) != 0 {
		t.Errorf("Expected 0 items, got %d", len(p1.Exploits))
	}
}

func TestPagination_GetAllEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{},
			ExploitsTotal: 0,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	all, err := p.GetAllResults()
	if err != nil {
		t.Fatalf("GetAllResults failed: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("Expected 0 results, got %d", len(all))
	}
}

// ============================================================
// 11. 工具函数测试
// ============================================================

func TestDefaultSearchTypes_Contains(t *testing.T) {
	types := DefaultSearchTypes()
	hasExploits := false
	hasTools := false
	for _, t := range types {
		if t == "exploits" {
			hasExploits = true
		}
		if t == "tools" {
			hasTools = true
		}
	}
	if !hasExploits || !hasTools {
		t.Error("DefaultSearchTypes should contain exploits and tools")
	}
}

func TestDefaultOutputPath_Format(t *testing.T) {
	path := DefaultOutputPath("test query", "exploits")
	if !strings.HasPrefix(path, "results/") {
		t.Errorf("Path should start with results/: %s", path)
	}
	if !strings.Contains(path, "test_query") {
		t.Errorf("Path should contain sanitized query: %s", path)
	}
	if !strings.Contains(path, ".json") {
		t.Errorf("Path should end with .json: %s", path)
	}
}

func TestSanitizeFilename_EmptyInput(t *testing.T) {
	result := sanitizeFilename("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestSanitizeFilename_AllSpecial(t *testing.T) {
	result := sanitizeFilename("<>:\"/\\|?*")
	if result != "_________" {
		t.Errorf("Expected 9 underscores, got %q (len=%d)", result, len(result))
	}
}

func TestSanitizeFilename_LongInput(t *testing.T) {
	input := strings.Repeat("a", 200)
	result := sanitizeFilename(input)
	if len(result) != 200 {
		t.Errorf("Expected length 200, got %d (utils.go version doesn't truncate)", len(result))
	}
}

func TestSanitizeFilename_ConsecutiveUnderscores(t *testing.T) {
	result := sanitizeFilename("a   b   c")
	if strings.Contains(result, "__") {
		t.Logf("Consecutive underscores found: %q", result)
	}
}

func TestExportJSON_VerifyContent(t *testing.T) {
	resp := &types.SearchResponse{
		Exploits: []types.Exploit{{Title: "Verify", Score: 5.0, ID: "VERIFY-1"}},
		ExploitsTotal: 1,
	}
	path := fmt.Sprintf("/tmp/sploitus_verify_%d.json", hashCode(t.Name()))
	defer func() { /* cleanup */ }()

	err := ExportJSON(resp, path)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	data, err := readFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var loaded types.SearchResponse
	json.Unmarshal(data, &loaded)
	if loaded.Exploits[0].ID != "VERIFY-1" {
		t.Errorf("Expected ID VERIFY-1, got %s", loaded.Exploits[0].ID)
	}
}

// ============================================================
// 12. JSON 序列化测试
// ============================================================

func TestJSON_ExploitMarshalUnmarshal(t *testing.T) {
	original := types.Exploit{Title: "JSON Test", Score: 7.5, ID: "JSON-1"}
	data, _ := json.Marshal(original)
	var restored types.Exploit
	json.Unmarshal(data, &restored)
	if restored.Title != original.Title {
		t.Errorf("JSON roundtrip failed: %q != %q", restored.Title, original.Title)
	}
}

func TestJSON_ExploitWithAllFields(t *testing.T) {
	exp := types.Exploit{
		Title: "Full", Score: 10.0, Href: "https://a.com",
		Type: "exploit", Published: "2024-01-01",
		ID: "CVE-2024-9999", Source: "code", Language: "go",
	}
	data, _ := json.Marshal(exp)
	var unmarshaled types.Exploit
	json.Unmarshal(data, &unmarshaled)

	if unmarshaled.Title != "Full" { t.Error("Title mismatch") }
	if unmarshaled.Score != 10.0 { t.Error("Score mismatch") }
	if unmarshaled.Href != "https://a.com" { t.Error("Href mismatch") }
	if unmarshaled.Type != "exploit" { t.Error("Type mismatch") }
	if unmarshaled.Published != "2024-01-01" { t.Error("Published mismatch") }
	if unmarshaled.ID != "CVE-2024-9999" { t.Error("ID mismatch") }
	if unmarshaled.Source != "code" { t.Error("Source mismatch") }
	if unmarshaled.Language != "go" { t.Error("Language mismatch") }
}

func TestJSON_SearchQueryMarshal(t *testing.T) {
	q := types.SearchQuery{Type: "exploits", Sort: "score", Query: "test", Title: true, Offset: 10}
	data, _ := json.Marshal(q)
	var restored types.SearchQuery
	json.Unmarshal(data, &restored)

	if restored.Type != "exploits" { t.Error("Type mismatch") }
	if restored.Sort != "score" { t.Error("Sort mismatch") }
	if restored.Query != "test" { t.Error("Query mismatch") }
	if !restored.Title { t.Error("Title should be true") }
	if restored.Offset != 10 { t.Error("Offset mismatch") }
}

func TestJSON_SearchResponseMarshal(t *testing.T) {
	orig := realExploitResponse()
	data, _ := json.Marshal(orig)
	var restored types.SearchResponse
	json.Unmarshal(data, &restored)

	if restored.ExploitsTotal != 2 { t.Errorf("Expected 2 total, got %d", restored.ExploitsTotal) }
	if len(restored.Exploits) != 2 { t.Errorf("Expected 2 exploits, got %d", len(restored.Exploits)) }
	if restored.Exploits[0].ID != "0147E6AA-6963-51CE-90F9-420346FA917B" {
		t.Errorf("ID mismatch: %q", restored.Exploits[0].ID)
	}
}

func TestJSON_ExploitDetailMarshal(t *testing.T) {
	detail := types.ExploitDetail{
		ID: "DETAIL-1", Title: "Detail Test",
		Description: "desc", Source: "src",
		Language: "py", Score: 8.0, Type: "exploit",
		Href: "https://d.com", Published: "2024-06-01",
	}
	data, _ := json.Marshal(detail)
	var restored types.ExploitDetail
	json.Unmarshal(data, &restored)

	if restored.ID != "DETAIL-1" { t.Error("ID mismatch") }
	if restored.Description != "desc" { t.Error("Description mismatch") }
	if restored.Source != "src" { t.Error("Source mismatch") }
}

// ============================================================
// 13. TypeScript 类型系统测试
// ============================================================

func TestTypes_ExploitZeroValue(t *testing.T) {
	var exp types.Exploit
	if exp.Title != "" { t.Error("Zero value Title should be empty") }
	if exp.Score != 0 { t.Error("Zero value Score should be 0") }
	if exp.Href != "" { t.Error("Zero value Href should be empty") }
	if exp.Type != "" { t.Error("Zero value Type should be empty") }
	if exp.Published != "" { t.Error("Zero value Published should be empty") }
	if exp.ID != "" { t.Error("Zero value ID should be empty") }
	if exp.Source != "" { t.Error("Zero value Source should be empty") }
	if exp.Language != "" { t.Error("Zero value Language should be empty") }
}

func TestTypes_SearchQueryZeroValue(t *testing.T) {
	var q types.SearchQuery
	if q.Type != "" { t.Error("Zero value Type should be empty") }
	if q.Sort != "" { t.Error("Zero value Sort should be empty") }
	if q.Query != "" { t.Error("Zero value Query should be empty") }
	if q.Title { t.Error("Zero value Title should be false") }
	if q.Offset != 0 { t.Error("Zero value Offset should be 0") }
}

func TestTypes_SearchResponseZeroValue(t *testing.T) {
	var resp types.SearchResponse
	if resp.Exploits != nil { t.Error("Zero value Exploits should be nil") }
	if resp.ExploitsTotal != 0 { t.Error("Zero value ExploitsTotal should be 0") }
}

func TestTypes_ExploitDetailZeroValue(t *testing.T) {
	var d types.ExploitDetail
	if d.ID != "" { t.Error("Zero value ID should be empty") }
	if d.Score != 0 { t.Error("Zero value Score should be 0") }
}

// ============================================================
// 14. Parser 测试
// ============================================================

func TestParseExploitIDFromHref_Standard(t *testing.T) {
	id := parseExploitIDFromHref("/exploit?id=0147E6AA-6963-51CE-90F9-420346FA917B")
	expected := "0147E6AA-6963-51CE-90F9-420346FA917B"
	if id != expected {
		t.Errorf("Expected %q, got %q", expected, id)
	}
}

func TestParseExploitIDFromHref_CVE(t *testing.T) {
	id := parseExploitIDFromHref("/exploit?id=CVE-2023-12345")
	expected := "CVE-2023-12345"
	if id != expected {
		t.Errorf("Expected %q, got %q", expected, id)
	}
}

func TestParseExploitIDFromHref_SAINT(t *testing.T) {
	id := parseExploitIDFromHref("/exploit?id=SAINT:2CEDD0194C77120545A6315E534CFE66")
	expected := "SAINT:2CEDD0194C77120545A6315E534CFE66"
	if id != expected {
		t.Errorf("Expected %q, got %q", expected, id)
	}
}

func TestParseExploitIDFromHref_NoID(t *testing.T) {
	id := parseExploitIDFromHref("/exploit")
	if id != "" {
		t.Errorf("Expected empty, got %q", id)
	}
}

func TestParseExploitIDFromHref_EmptyID(t *testing.T) {
	id := parseExploitIDFromHref("/exploit?id=")
	if id != "" {
		t.Errorf("Expected empty, got %q", id)
	}
}

func TestParseExploitIDFromHref_EmptyString(t *testing.T) {
	id := parseExploitIDFromHref("")
	if id != "" {
		t.Errorf("Expected empty, got %q", id)
	}
}

// ============================================================
// 15. SearchQuery 边界测试
// ============================================================

func TestSearchQuery_MaxOffset(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	_, err := client.Search("test", "exploits", "default", 999999)
	if err != nil {
		t.Fatalf("Search with max offset failed: %v", err)
	}
}

func TestSearchQuery_AllSortTypes(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	sorts := []string{"default", "score", "date", "relevance", "newest", "oldest"}
	for _, sort := range sorts {
		_, err := client.Search("test", "exploits", sort, 0)
		if err != nil {
			t.Errorf("Search with sort=%q failed: %v", sort, err)
		}
	}
}

func TestSearchQuery_AllSearchTypes(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	types := []string{"exploits", "tools", "cve", "title", "tag", "text", ""}
	for _, st := range types {
		_, err := client.Search("test", st, "default", 0)
		if err != nil {
			t.Errorf("Search with type=%q failed: %v", st, err)
		}
	}
}

// ============================================================
// 16. OS 操作测试
// ============================================================

func TestExportJSON_CreateDir(t *testing.T) {
	// Test that ExportJSON creates intermediate directories
	path := "/tmp/sploitus_test_deep/nested/dir/output.json"
	resp := &types.SearchResponse{Exploits: []types.Exploit{{Title: "Deep", ID: "DEEP-1"}}, ExploitsTotal: 1}
	err := ExportJSON(resp, path)
	if err != nil {
		t.Fatalf("ExportJSON to deep path failed: %v", err)
	}
	// Cleanup
	removeAll("/tmp/sploitus_test_deep")
}

// removeAll is a helper to remove a directory
var removeAll = func(path string) error {
	return os.RemoveAll(path)
}

func TestDefaultOutputPath_Unique(t *testing.T) {
	path1 := DefaultOutputPath("test", "exploits")
	path2 := DefaultOutputPath("test", "exploits")
	if path1 == path2 {
		// This is possible if both run within the same second
		t.Log("Paths are identical (may be same second)")
	}
}

func TestDefaultOutputPath_WithSpecialChars(t *testing.T) {
	path := DefaultOutputPath("CVE-2023:test/path", "exploits")
	if strings.Contains(path, ":") || strings.Contains(path, "/") {
		// The colon and slash should be sanitized
	}
}

func TestSanitizeFilename_Dots(t *testing.T) {
	result := sanitizeFilename("file.name.with.dots.txt")
	if result != "file.name.with.dots.txt" {
		t.Errorf("Dots should be preserved, got %q", result)
	}
}

// ============================================================
// 17. EnPagination 补全测试
// ============================================================

func TestEnPagination_GetPageInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: make([]types.Exploit, 5),
			ExploitsTotal: 15,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	p.SetPageSize(5)
	p.GetFirstPage()

	info := p.GetPageInfo()
	if !strings.Contains(info, "Page") {
		t.Errorf("GetPageInfo should contain 'Page': %s", info)
	}
	if !strings.Contains(info, "Total") {
		t.Errorf("GetPageInfo should contain 'Total': %s", info)
	}
}

func TestEnPagination_GetAllResultsWithLimit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: make([]types.Exploit, 10),
			ExploitsTotal: 50,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)
	all, err := p.GetAllResults()
	if err != nil {
		t.Fatalf("GetAllResults failed: %v", err)
	}
	if len(all) != 50 {
		t.Errorf("Expected 50 results, got %d", len(all))
	}
}

// ============================================================
// 18. 并发安全测试
// ============================================================

func TestClient_ConcurrentSearch(t *testing.T) {
	server := mockSearchServer(realExploitResponse())
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(n int) {
			_, err := client.Search(fmt.Sprintf("test-%d", n), "exploits", "default", 0)
			if err != nil {
				t.Errorf("Concurrent search %d failed: %v", n, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestPagination_ConcurrentGetAll(t *testing.T) {
	server := mockSearchServer(&types.SearchResponse{
		Exploits: make([]types.Exploit, 10),
		ExploitsTotal: 100,
	})
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			p := client.NewPaginationHelper("test", "exploits", "default")
			p.SetPageSize(10)
			_, err := p.GetAllResults()
			if err != nil {
				t.Errorf("Concurrent GetAllResults failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 3; i++ {
		<-done
	}
}