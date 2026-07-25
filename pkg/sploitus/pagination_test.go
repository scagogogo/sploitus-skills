package sploitus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

func TestPaginationHelper(t *testing.T) {
	// 设置测试数据
	totalItems := 57   // 总条目数
	pageSize := 10     // 页面大小
	expectedPages := 6 // 预期总页数（向上取整）

	// 创建模拟服务器处理分页请求
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析请求体
		var query types.SearchQuery
		err := json.NewDecoder(r.Body).Decode(&query)
		if err != nil {
			t.Errorf("解析请求体失败: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 获取偏移量
		offset := query.Offset

		// 计算当前页应该返回的结果数
		itemsInPage := pageSize
		if offset+pageSize > totalItems {
			itemsInPage = totalItems - offset
			if itemsInPage < 0 {
				itemsInPage = 0
			}
		}

		// 创建模拟响应
		exploits := make([]types.Exploit, itemsInPage)
		for i := 0; i < itemsInPage; i++ {
			itemIndex := offset + i
			exploits[i] = types.Exploit{
				Title:     "测试漏洞 " + strconv.Itoa(itemIndex),
				Score:     7.5,
				Href:      "https://example.com/" + strconv.Itoa(itemIndex),
				Type:      "test",
				Published: "2023-01-01",
				ID:        "TEST-" + strconv.Itoa(itemIndex),
				Source:    "test-source",
				Language:  "test",
			}
		}

		response := types.SearchResponse{
			Exploits:      exploits,
			ExploitsTotal: totalItems,
		}

		// 返回JSON响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端和分页助手
	client := NewClient()
	client.BaseURL = server.URL

	paginator := client.NewPaginationHelper("测试查询", "exploits", "default")
	paginator.SetPageSize(pageSize)

	// 测试 GetFirstPage
	firstPage, err := paginator.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage 失败: %v", err)
	}

	if len(firstPage.Exploits) != pageSize {
		t.Errorf("第一页应该有 %d 个结果，实际有 %d 个", pageSize, len(firstPage.Exploits))
	}

	if firstPage.ExploitsTotal != totalItems {
		t.Errorf("总条目数应该是 %d，实际是 %d", totalItems, firstPage.ExploitsTotal)
	}

	// 测试 GetNextPage
	secondPage, err := paginator.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage 失败: %v", err)
	}

	if len(secondPage.Exploits) != pageSize {
		t.Errorf("第二页应该有 %d 个结果，实际有 %d 个", pageSize, len(secondPage.Exploits))
	}

	if paginator.GetCurrentPosition() != pageSize {
		t.Errorf("当前位置应该是 %d，实际是 %d", pageSize, paginator.GetCurrentPosition())
	}

	// 测试 GetPage
	thirdPage, err := paginator.GetPage(3)
	if err != nil {
		t.Fatalf("GetPage 失败: %v", err)
	}

	if len(thirdPage.Exploits) != pageSize {
		t.Errorf("第三页应该有 %d 个结果，实际有 %d 个", pageSize, len(thirdPage.Exploits))
	}

	if paginator.GetCurrentPosition() != pageSize*2 {
		t.Errorf("当前位置应该是 %d，实际是 %d", pageSize*2, paginator.GetCurrentPosition())
	}

	// 测试 GetPage 超出范围
	lastPage, err := paginator.GetPage(6)
	if err != nil {
		t.Fatalf("GetPage 失败: %v", err)
	}

	expectedLastPageItems := totalItems - pageSize*5
	if len(lastPage.Exploits) != expectedLastPageItems {
		t.Errorf("最后一页应该有 %d 个结果，实际有 %d 个", expectedLastPageItems, len(lastPage.Exploits))
	}

	// 测试超出范围的页面
	tooFarPage, err := paginator.GetPage(7)
	if err != nil {
		t.Fatalf("GetPage 失败: %v", err)
	}

	if len(tooFarPage.Exploits) != 0 {
		t.Errorf("超出范围的页面应该有 0 个结果，实际有 %d 个", len(tooFarPage.Exploits))
	}

	// 测试 GetTotalPages
	totalPages, err := paginator.GetTotalPages()
	if err != nil {
		t.Fatalf("GetTotalPages 失败: %v", err)
	}

	if totalPages != expectedPages {
		t.Errorf("总页数应该是 %d，实际是 %d", expectedPages, totalPages)
	}

	// 测试 Reset
	paginator.Reset()
	if paginator.GetCurrentPosition() != 0 {
		t.Errorf("Reset 后当前位置应该是 0，实际是 %d", paginator.GetCurrentPosition())
	}

	// 测试页面信息
	pageInfo := paginator.GetPageInfo()
	if pageInfo == "" {
		t.Error("GetPageInfo 不应该返回空字符串")
	}

	// 测试获取所有结果
	allResults, err := paginator.GetAllResults()
	if err != nil {
		t.Fatalf("GetAllResults 失败: %v", err)
	}

	if len(allResults) != totalItems {
		t.Errorf("所有结果应该有 %d 个，实际有 %d 个", totalItems, len(allResults))
	}
}

func TestEnPaginationHelper(t *testing.T) {
	// 设置测试数据
	totalItems := 37   // 总条目数
	pageSize := 5      // 页面大小
	expectedPages := 8 // 预期总页数（向上取整）

	// 创建模拟服务器处理分页请求
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析请求体
		var query types.SearchQuery
		err := json.NewDecoder(r.Body).Decode(&query)
		if err != nil {
			t.Errorf("Failed to parse request body: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 获取偏移量
		offset := query.Offset

		// 计算当前页应该返回的结果数
		itemsInPage := pageSize
		if offset+pageSize > totalItems {
			itemsInPage = totalItems - offset
			if itemsInPage < 0 {
				itemsInPage = 0
			}
		}

		// 创建模拟响应
		exploits := make([]types.Exploit, itemsInPage)
		for i := 0; i < itemsInPage; i++ {
			itemIndex := offset + i
			exploits[i] = types.Exploit{
				Title:     "Test Exploit " + strconv.Itoa(itemIndex),
				Score:     8.5,
				Href:      "https://example.com/" + strconv.Itoa(itemIndex),
				Type:      "test",
				Published: "2023-01-01",
				ID:        "TEST-" + strconv.Itoa(itemIndex),
				Source:    "test-source",
				Language:  "test",
			}
		}

		response := types.SearchResponse{
			Exploits:      exploits,
			ExploitsTotal: totalItems,
		}

		// 返回JSON响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端和分页助手
	client := NewClient()
	client.BaseURL = server.URL

	paginator := client.NewEnPaginationHelper("test query", "exploits", "default")
	paginator.SetPageSize(pageSize)

	// 测试 GetFirstPage
	firstPage, err := paginator.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	if len(firstPage.Exploits) != pageSize {
		t.Errorf("First page should have %d results, got %d", pageSize, len(firstPage.Exploits))
	}

	// 测试 GetTotalPages
	totalPages, err := paginator.GetTotalPages()
	if err != nil {
		t.Fatalf("GetTotalPages failed: %v", err)
	}

	if totalPages != expectedPages {
		t.Errorf("Total pages should be %d, got %d", expectedPages, totalPages)
	}

	// 测试 GetPageInfo
	pageInfo := paginator.GetPageInfo()
	if pageInfo == "" {
		t.Error("GetPageInfo should not return an empty string")
	}

	// 测试剩余方法覆盖
	paginator2 := client.NewEnPaginationHelper("test query", "exploits", "default")
	paginator2.SetPageSize(pageSize)

	// GetFirstPage for paginator2
	_, err = paginator2.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	// HasMore should return true (total not yet known, position 0)
	if !paginator2.HasMore() {
		t.Error("HasMore should return true when total is unknown")
	}

	// GetPageSize
	if paginator2.GetPageSize() != pageSize {
		t.Errorf("GetPageSize should be %d, got %d", pageSize, paginator2.GetPageSize())
	}

	// GetCurrentPage
	if paginator2.GetCurrentPage() != 1 {
		t.Errorf("GetCurrentPage should be 1, got %d", paginator2.GetCurrentPage())
	}

	// GetNextPage
	nextPage, err := paginator2.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}
	if len(nextPage.Exploits) == 0 {
		t.Error("Expected at least 1 exploit in next page")
	}

	// GetPage
	page3, err := paginator2.GetPage(3)
	if err != nil {
		t.Fatalf("GetPage(3) failed: %v", err)
	}
	if len(page3.Exploits) == 0 {
		t.Error("Expected at least 1 exploit in page 3")
	}

	// GetCurrentPosition
	if paginator2.GetCurrentPosition() <= 0 {
		t.Errorf("GetCurrentPosition should be > 0, got %d", paginator2.GetCurrentPosition())
	}

	// Reset
	paginator2.Reset()
	if paginator2.GetCurrentPosition() != 0 {
		t.Errorf("After Reset, GetCurrentPosition should be 0, got %d", paginator2.GetCurrentPosition())
	}

	// GetAllResults (should work from scratch)
	client2 := NewClient()
	client2.BaseURL = server.URL
	paginator3 := client2.NewEnPaginationHelper("test query", "exploits", "default")
	paginator3.SetPageSize(pageSize)

	allResults, err := paginator3.GetAllResults()
	if err != nil {
		t.Fatalf("GetAllResults failed: %v", err)
	}
	if len(allResults) != totalItems {
		t.Errorf("GetAllResults should return %d items, got %d", totalItems, len(allResults))
	}

	// GetPageInfo after having data
	info := paginator3.GetPageInfo()
	if info == "" {
		t.Error("GetPageInfo should not be empty")
	}
}

func TestPaginationHasMore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query types.SearchQuery
		json.NewDecoder(r.Body).Decode(&query)
		offset := query.Offset
		if offset >= 1 {
			json.NewEncoder(w).Encode(types.SearchResponse{
				Exploits:      []types.Exploit{},
				ExploitsTotal: 1,
			})
			return
		}
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0, Href: "https://example.com",
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	// HasMore should return true when total is unknown
	if !p.HasMore() {
		t.Error("HasMore should return true when total is unknown")
	}

	// Fetch first page
	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	if !p.HasMore() {
		t.Error("HasMore should return true when currentPos < totalItems")
	}

	// GetNextPage should advance past the end
	_, err = p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}
	if p.HasMore() {
		t.Error("HasMore should return false when currentPos >= totalItems")
	}
}

func TestPaginationGetPageSize(t *testing.T) {
	client := NewClient()
	p := client.NewPaginationHelper("test", "exploits", "default")
	if p.GetPageSize() != DefaultPageSize {
		t.Errorf("Default GetPageSize should be %d, got %d", DefaultPageSize, p.GetPageSize())
	}
	p.SetPageSize(5)
	if p.GetPageSize() != 5 {
		t.Errorf("GetPageSize after SetPageSize(5) should be 5, got %d", p.GetPageSize())
	}
}

func TestPaginationGetPage_Invalid(t *testing.T) {
	client := NewClient()
	p := client.NewPaginationHelper("test", "exploits", "default")
	_, err := p.GetPage(0)
	if err == nil {
		t.Error("Expected error for page 0, got nil")
	}
}

func TestPaginationGetPageInfo_UnknownTotal(t *testing.T) {
	client := NewClient()
	p := client.NewPaginationHelper("test", "exploits", "default")
	info := p.GetPageInfo()
	if !strings.Contains(info, "未知") && !strings.Contains(info, "unknown") {
		t.Errorf("GetPageInfo with unknown total should indicate unknown, got: %s", info)
	}
}

func TestPaginationGetNextPage_NoMore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query types.SearchQuery
		json.NewDecoder(r.Body).Decode(&query)
		offset := query.Offset
		if offset >= 1 {
			json.NewEncoder(w).Encode(types.SearchResponse{
				Exploits:      []types.Exploit{},
				ExploitsTotal: 1,
			})
			return
		}
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0,
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	// Get first page
	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	// GetNextPage — totalItems=1, pageSize=10, so we'll advance past end
	np, err := p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}
	if len(np.Exploits) != 0 {
		t.Errorf("GetNextPage should return empty results when at end, got %d items", len(np.Exploits))
	}
}

func TestEnPaginationGetPageInfo_UnknownTotal(t *testing.T) {
	client := NewClient()
	p := client.NewEnPaginationHelper("test", "exploits", "default")
	info := p.GetPageInfo()
	if !strings.Contains(info, "unknown") {
		t.Errorf("GetPageInfo with unknown total should indicate unknown, got: %s", info)
	}
}

func TestPaginationGetTotalPages_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	_, err := p.GetTotalPages()
	if err == nil {
		t.Error("Expected error when server returns 500, got nil")
	}
}

func TestPaginationGetAllResults_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	_, err := p.GetAllResults()
	if err == nil {
		t.Error("Expected error when server returns 500, got nil")
	}
}

func TestPaginationGetNextPage_Error(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(types.SearchResponse{
				Exploits: []types.Exploit{{
					Title: "test", Score: 5.0, Href: "https://example.com",
					Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
					Source: "test", Language: "test",
				}},
				ExploitsTotal: 100,
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	_, err = p.GetNextPage()
	if err == nil {
		t.Error("Expected error when server returns 500 on GetNextPage, got nil")
	}
}

func TestPaginationGetTotalPages_WithCachedTotal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0, Href: "https://example.com",
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 25,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	// First call should fetch and cache
	pages, err := p.GetTotalPages()
	if err != nil {
		t.Fatalf("GetTotalPages failed: %v", err)
	}
	if pages != 3 {
		t.Errorf("Expected 3 pages for 25 items at page size 10, got %d", pages)
	}

	// Second call should use cached total
	pages, err = p.GetTotalPages()
	if err != nil {
		t.Fatalf("GetTotalPages (cached) failed: %v", err)
	}
	if pages != 3 {
		t.Errorf("Expected 3 pages (cached), got %d", pages)
	}
}

func TestPaginationGetPage_WithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	_, err := p.GetPage(1)
	if err == nil {
		t.Error("Expected error when server returns 500 on GetPage, got nil")
	}
}

func TestEnPaginationHasMore_TotalUnknown(t *testing.T) {
	client := NewClient()
	p := client.NewEnPaginationHelper("test", "exploits", "default")
	if !p.HasMore() {
		t.Error("HasMore should return true when total is unknown")
	}
}

func TestEnPaginationGetNextPage_Error(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(types.SearchResponse{
				Exploits: []types.Exploit{{
					Title: "test", Score: 5.0, Href: "https://example.com",
					Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
					Source: "test", Language: "test",
				}},
				ExploitsTotal: 100,
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	_, err = p.GetNextPage()
	if err == nil {
		t.Error("Expected error when server returns 500 on GetNextPage, got nil")
	}
}

func TestEnPaginationGetPage_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	_, err := p.GetPage(1)
	if err == nil {
		t.Error("Expected error when server returns 500 on GetPage, got nil")
	}
}

func TestEnPaginationGetAllResults_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	_, err := p.GetAllResults()
	if err == nil {
		t.Error("Expected error when server returns 500 on GetAllResults, got nil")
	}
}

func TestEnPaginationGetTotalPages_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	_, err := p.GetTotalPages()
	if err == nil {
		t.Error("Expected error when server returns 500 on GetTotalPages, got nil")
	}
}

func TestEnPaginationGetPage_Invalid(t *testing.T) {
	client := NewClient()
	p := client.NewEnPaginationHelper("test", "exploits", "default")
	_, err := p.GetPage(0)
	if err == nil {
		t.Error("Expected error for page 0, got nil")
	}
}

func TestExportJSON_PermissionError(t *testing.T) {
	response := &types.SearchResponse{
		Exploits: []types.Exploit{
			{Title: "test", Score: 5.0, Type: "exploits", ID: "TEST-1"},
		},
		ExploitsTotal: 1,
	}
	// Use a path that's likely to fail (root-owned directory, or a file path with null bytes)
	err := ExportJSON(response, "/root/forbidden_path/file.json")
	// We expect an error (either permission denied or no such directory)
	if err == nil {
		t.Error("Expected error for invalid output path, got nil")
	}
}
func TestEnPaginationHasMore_False(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0, Href: "https://example.com",
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	// GetFirstPage
	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	// HasMore should be true (currentPos=0, totalItems=1)
	if !p.HasMore() {
		t.Error("HasMore should be true when currentPos < totalItems")
	}

	// GetNextPage to advance past end
	_, err = p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}

	// Now HasMore should be false (currentPos=10, totalItems=1)
	if p.HasMore() {
		t.Error("HasMore should be false when currentPos >= totalItems")
	}
}

func TestPaginationGetNextPage_AlreadyAtEnd(t *testing.T) {
	// When totalItems >= 0 and currentPos >= totalItems, GetNextPage should return empty
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0, Href: "https://example.com",
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	// Get first page to set totalItems
	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	// currentPos=0, totalItems=1, so advance past end
	_, err = p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}
	// currentPos=10, totalItems=1, so now at end
	// This should trigger the early return branch
	finalPage, err := p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage (at end) failed: %v", err)
	}
	if len(finalPage.Exploits) != 0 {
		t.Errorf("Expected empty page when at end, got %d items", len(finalPage.Exploits))
	}
}

func TestEnPaginationGetNextPage_AlreadyAtEnd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(types.SearchResponse{
			Exploits: []types.Exploit{{
				Title: "test", Score: 5.0, Href: "https://example.com",
				Type: "exploits", Published: "2023-01-01", ID: "TEST-0",
				Source: "test", Language: "test",
			}},
			ExploitsTotal: 1,
		})
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL

	p := client.NewEnPaginationHelper("test", "exploits", "default")
	p.SetPageSize(10)

	_, err := p.GetFirstPage()
	if err != nil {
		t.Fatalf("GetFirstPage failed: %v", err)
	}

	_, err = p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage failed: %v", err)
	}

	// Now at end - should trigger early return
	finalPage, err := p.GetNextPage()
	if err != nil {
		t.Fatalf("GetNextPage (at end) failed: %v", err)
	}
	if len(finalPage.Exploits) != 0 {
		t.Errorf("Expected empty page when at end, got %d items", len(finalPage.Exploits))
	}
}
