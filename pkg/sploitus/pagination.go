package sploitus

import (
	"fmt"

	"github.com/scagogogo/sploitus-skills/pkg/types"
)

// DefaultPageSize 是每页默认的结果数量
const DefaultPageSize = 20

// PaginationHelper 提供分页功能的辅助结构体
type PaginationHelper struct {
	client     *Client
	query      string
	searchType string
	sort       string
	pageSize   int
	totalItems int
	currentPos int
}

// NewPaginationHelper 创建一个新的分页辅助器
func (c *Client) NewPaginationHelper(query string, searchType string, sort string) *PaginationHelper {
	return &PaginationHelper{
		client:     c,
		query:      query,
		searchType: searchType,
		sort:       sort,
		pageSize:   DefaultPageSize,
		totalItems: -1, // 表示尚未获取总数
		currentPos: 0,
	}
}

// SetPageSize 设置每页结果数量
func (p *PaginationHelper) SetPageSize(pageSize int) *PaginationHelper {
	if pageSize > 0 {
		p.pageSize = pageSize
	}
	return p
}

// GetFirstPage 获取第一页结果
func (p *PaginationHelper) GetFirstPage() (*types.SearchResponse, error) {
	p.currentPos = 0
	return p.client.Search(p.query, p.searchType, p.sort, p.currentPos)
}

// GetNextPage 获取下一页结果
func (p *PaginationHelper) GetNextPage() (*types.SearchResponse, error) {
	if p.totalItems >= 0 && p.currentPos >= p.totalItems {
		return &types.SearchResponse{
			Exploits:      []types.Exploit{},
			ExploitsTotal: p.totalItems,
		}, nil
	}

	p.currentPos += p.pageSize
	resp, err := p.client.Search(p.query, p.searchType, p.sort, p.currentPos)
	if err != nil {
		return nil, err
	}

	// 更新总数
	p.totalItems = resp.ExploitsTotal

	return resp, nil
}

// GetPage 获取指定页码的结果（页码从1开始）
func (p *PaginationHelper) GetPage(pageNum int) (*types.SearchResponse, error) {
	if pageNum < 1 {
		return nil, fmt.Errorf("页码必须大于等于1")
	}

	offset := (pageNum - 1) * p.pageSize
	p.currentPos = offset

	return p.client.Search(p.query, p.searchType, p.sort, offset)
}

// HasMore 检查是否还有更多结果
func (p *PaginationHelper) HasMore() bool {
	// 如果还没有获取总数，假设还有更多
	if p.totalItems < 0 {
		return true
	}

	return p.currentPos < p.totalItems
}

// GetTotalPages 获取总页数
func (p *PaginationHelper) GetTotalPages() (int, error) {
	// 如果尚未获取总数，先获取第一页
	if p.totalItems < 0 {
		resp, err := p.GetFirstPage()
		if err != nil {
			return 0, err
		}
		p.totalItems = resp.ExploitsTotal
	}

	// 计算总页数（向上取整）
	totalPages := (p.totalItems + p.pageSize - 1) / p.pageSize
	return totalPages, nil
}

// GetAllResults 获取所有结果
func (p *PaginationHelper) GetAllResults() ([]types.Exploit, error) {
	// 重置位置确保从头开始
	p.currentPos = 0

	var allExploits []types.Exploit

	for {
		resp, err := p.client.Search(p.query, p.searchType, p.sort, p.currentPos)
		if err != nil {
			return nil, err
		}

		// 更新总数
		p.totalItems = resp.ExploitsTotal

		// 添加结果
		allExploits = append(allExploits, resp.Exploits...)

		// 检查是否所有结果都已获取
		if len(allExploits) >= p.totalItems || len(resp.Exploits) == 0 {
			break
		}

		// 移动到下一页
		p.currentPos += p.pageSize
	}

	return allExploits, nil
}

// Reset 重置分页状态
func (p *PaginationHelper) Reset() {
	p.currentPos = 0
}

// GetCurrentPosition 获取当前位置
func (p *PaginationHelper) GetCurrentPosition() int {
	return p.currentPos
}

// GetCurrentPage 获取当前页码（从1开始）
func (p *PaginationHelper) GetCurrentPage() int {
	return (p.currentPos / p.pageSize) + 1
}

// GetPageSize 获取当前设置的页大小
func (p *PaginationHelper) GetPageSize() int {
	return p.pageSize
}

// GetPageInfo 获取当前页信息的简明描述
func (p *PaginationHelper) GetPageInfo() string {
	if p.totalItems < 0 {
		return fmt.Sprintf("当前位置: %d, 页大小: %d, 总数: 未知", p.currentPos, p.pageSize)
	}

	currentPage := p.GetCurrentPage()
	totalPages, _ := p.GetTotalPages()

	return fmt.Sprintf("第 %d/%d 页, 总共 %d 条结果", currentPage, totalPages, p.totalItems)
}
