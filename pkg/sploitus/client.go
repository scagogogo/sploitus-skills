package sploitus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/scagogogo/sploitus-crawler/pkg/types"
)

const (
	DefaultBaseURL = "https://sploitus.com"
	SearchEndpoint = "/search"
	DefaultTimeout = 30 * time.Second
)

// Client represents a Sploitus API client
type Client struct {
	BaseURL        string
	HTTPClient     *http.Client
	UserAgent      string
	Cookies        []*http.Cookie
	AcceptLanguage string
}

// NewClient creates a new Sploitus API client with default settings
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		AcceptLanguage: "zh-CN,zh;q=0.9",
	}
}

// NewClientWithProxy creates a new Sploitus API client with a proxy
func NewClientWithProxy(proxyURL string) (*Client, error) {
	// Parse the proxy URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Create a transport with the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	// Create an HTTP client with the custom transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   DefaultTimeout,
	}

	return &Client{
		BaseURL:        DefaultBaseURL,
		HTTPClient:     httpClient,
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
		AcceptLanguage: "zh-CN,zh;q=0.9",
	}, nil
}

// SetProxy sets or changes the proxy for an existing client
func (c *Client) SetProxy(proxyURL string) error {
	// Parse the proxy URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Get the existing transport or create a new one
	var transport *http.Transport
	if c.HTTPClient.Transport == nil {
		transport = &http.Transport{}
	} else {
		// Try to cast existing transport to http.Transport
		var ok bool
		transport, ok = c.HTTPClient.Transport.(*http.Transport)
		if !ok {
			// If it's a different type of transport, create a new one
			transport = &http.Transport{}
		}
	}

	// Set the proxy
	transport.Proxy = http.ProxyURL(proxy)

	// Update the client's transport
	c.HTTPClient.Transport = transport

	return nil
}

// SetCookies sets the cookies for the client
func (c *Client) SetCookies(cookieStr string) error {
	if cookieStr == "" {
		return nil
	}

	header := http.Header{}
	header.Add("Cookie", cookieStr)
	request := http.Request{Header: header}

	c.Cookies = request.Cookies()
	return nil
}

// Search performs a search query on the Sploitus API
func (c *Client) Search(query string, searchType string, sort string, offset int) (*types.SearchResponse, error) {
	searchQuery := types.SearchQuery{
		Type:   searchType,
		Sort:   sort,
		Query:  query,
		Title:  true,
		Offset: offset,
	}

	return c.SearchWithQuery(&searchQuery)
}

// SearchWithQuery performs a search with a custom query object
func (c *Client) SearchWithQuery(searchQuery *types.SearchQuery) (*types.SearchResponse, error) {
	jsonData, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// 打印实际发送的请求数据
	fmt.Printf("DEBUG - 请求数据: %s\n", string(jsonData))

	url := c.BaseURL + SearchEndpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置更多模拟真实浏览器的请求头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", c.AcceptLanguage)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", c.BaseURL)
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", c.BaseURL+"/?query="+searchQuery.Query)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Sec-Ch-Ua", "\"Chromium\";v=\"134\", \"Not:A-Brand\";v=\"24\", \"Google Chrome\";v=\"134\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 添加Cookies
	for _, cookie := range c.Cookies {
		req.AddCookie(cookie)
		fmt.Printf("DEBUG - 添加Cookie: %s=%s\n", cookie.Name, cookie.Value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取错误响应
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned non-200 status code: %d, response: %s", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 打印实际接收的响应数据
	fmt.Printf("DEBUG - 响应数据: %s\n", string(body))

	var searchResponse types.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResponse, nil
}
