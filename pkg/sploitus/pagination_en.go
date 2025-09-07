package sploitus

import (
	"fmt"

	"github.com/scagogogo/sploitus-crawler/pkg/types"
)

// EnPaginationHelper provides helper functions for pagination in English
type EnPaginationHelper struct {
	client     *Client
	query      string
	searchType string
	sort       string
	pageSize   int
	totalItems int
	currentPos int
}

// NewEnPaginationHelper creates a new pagination helper with English interface
func (c *Client) NewEnPaginationHelper(query string, searchType string, sort string) *EnPaginationHelper {
	return &EnPaginationHelper{
		client:     c,
		query:      query,
		searchType: searchType,
		sort:       sort,
		pageSize:   DefaultPageSize,
		totalItems: -1, // indicates total count is not fetched yet
		currentPos: 0,
	}
}

// SetPageSize sets the number of results per page
func (p *EnPaginationHelper) SetPageSize(pageSize int) *EnPaginationHelper {
	if pageSize > 0 {
		p.pageSize = pageSize
	}
	return p
}

// GetFirstPage retrieves the first page of results
func (p *EnPaginationHelper) GetFirstPage() (*types.SearchResponse, error) {
	p.currentPos = 0
	return p.client.Search(p.query, p.searchType, p.sort, p.currentPos)
}

// GetNextPage retrieves the next page of results
func (p *EnPaginationHelper) GetNextPage() (*types.SearchResponse, error) {
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

	// Update total count
	p.totalItems = resp.ExploitsTotal

	return resp, nil
}

// GetPage retrieves a specific page of results (page numbers start from 1)
func (p *EnPaginationHelper) GetPage(pageNum int) (*types.SearchResponse, error) {
	if pageNum < 1 {
		return nil, fmt.Errorf("page number must be at least 1")
	}

	offset := (pageNum - 1) * p.pageSize
	p.currentPos = offset

	return p.client.Search(p.query, p.searchType, p.sort, offset)
}

// HasMore checks if there are more results available
func (p *EnPaginationHelper) HasMore() bool {
	// If we haven't fetched the total count yet, assume there are more
	if p.totalItems < 0 {
		return true
	}

	return p.currentPos < p.totalItems
}

// GetTotalPages returns the total number of pages
func (p *EnPaginationHelper) GetTotalPages() (int, error) {
	// If total count hasn't been fetched yet, get the first page
	if p.totalItems < 0 {
		resp, err := p.GetFirstPage()
		if err != nil {
			return 0, err
		}
		p.totalItems = resp.ExploitsTotal
	}

	// Calculate total pages (round up)
	totalPages := (p.totalItems + p.pageSize - 1) / p.pageSize
	return totalPages, nil
}

// GetAllResults retrieves all available results
func (p *EnPaginationHelper) GetAllResults() ([]types.Exploit, error) {
	// Reset position to ensure we start from the beginning
	p.currentPos = 0

	var allExploits []types.Exploit

	for {
		resp, err := p.client.Search(p.query, p.searchType, p.sort, p.currentPos)
		if err != nil {
			return nil, err
		}

		// Update total count
		p.totalItems = resp.ExploitsTotal

		// Add results
		allExploits = append(allExploits, resp.Exploits...)

		// Check if all results have been retrieved
		if len(allExploits) >= p.totalItems || len(resp.Exploits) == 0 {
			break
		}

		// Move to next page
		p.currentPos += p.pageSize
	}

	return allExploits, nil
}

// Reset resets the pagination state
func (p *EnPaginationHelper) Reset() {
	p.currentPos = 0
}

// GetCurrentPosition returns the current position (offset)
func (p *EnPaginationHelper) GetCurrentPosition() int {
	return p.currentPos
}

// GetCurrentPage returns the current page number (starting from 1)
func (p *EnPaginationHelper) GetCurrentPage() int {
	return (p.currentPos / p.pageSize) + 1
}

// GetPageSize returns the current page size setting
func (p *EnPaginationHelper) GetPageSize() int {
	return p.pageSize
}

// GetPageInfo returns a concise description of the current page information
func (p *EnPaginationHelper) GetPageInfo() string {
	if p.totalItems < 0 {
		return fmt.Sprintf("Position: %d, Page size: %d, Total: unknown", p.currentPos, p.pageSize)
	}

	currentPage := p.GetCurrentPage()
	totalPages, _ := p.GetTotalPages()

	return fmt.Sprintf("Page %d of %d, Total %d items", currentPage, totalPages, p.totalItems)
}
