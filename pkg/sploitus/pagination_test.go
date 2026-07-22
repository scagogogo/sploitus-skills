package sploitus

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
}
