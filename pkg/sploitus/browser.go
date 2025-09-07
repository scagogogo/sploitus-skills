package sploitus

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/scagogogo/sploitus-crawler/pkg/types"
)

// BrowserSearcher 使用真实浏览器执行搜索
type BrowserSearcher struct {
	launcher *launcher.Launcher
	browser  *rod.Browser
	debug    bool
}

// NewBrowserSearcher 创建一个新的浏览器搜索器
func NewBrowserSearcher(debug bool) (*BrowserSearcher, error) {
	// 设置无头浏览器
	l := launcher.New().
		Headless(true).  // 无头模式
		Devtools(false). // 关闭DevTools
		UserDataDir(""). // 不使用用户数据目录
		NoSandbox(true). // 禁用沙箱（在某些环境中可能需要）
		Proxy("")        // 不使用代理

	if debug {
		l = l.Headless(false) // 调试模式下显示浏览器窗口
	}

	url := l.MustLaunch()

	// 创建浏览器
	browser := rod.New().
		ControlURL(url).
		Trace(debug).
		SlowMotion(500 * time.Millisecond) // 调试时延迟操作

	// 连接浏览器
	err := browser.Connect()
	if err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %w", err)
	}

	return &BrowserSearcher{
		launcher: l,
		browser:  browser,
		debug:    debug,
	}, nil
}

// Close 关闭浏览器
func (bs *BrowserSearcher) Close() error {
	return bs.browser.Close()
}

// Search 使用浏览器执行搜索并获取结果
func (bs *BrowserSearcher) Search(query string, searchType string, sort string, offset int) (*types.SearchResponse, error) {
	// 创建页面
	page, err := bs.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	// 先访问主页以获取CloudFlare clearance
	if err := page.Navigate(DefaultBaseURL); err != nil {
		return nil, fmt.Errorf("导航到Sploitus主页失败: %w", err)
	}

	// 等待页面加载完成并通过CloudFlare检查
	if err := page.WaitIdle(30 * time.Second); err != nil {
		return nil, fmt.Errorf("等待页面加载完成失败: %w", err)
	}

	// 准备搜索查询
	searchQuery := types.SearchQuery{
		Type:   searchType,
		Sort:   sort,
		Query:  query,
		Title:  true,
		Offset: offset,
	}

	jsonData, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("序列化搜索查询失败: %w", err)
	}

	// 设置请求拦截器来捕获API响应
	router := page.HijackRequests()
	defer router.Stop()

	// 创建一个通道来接收API响应
	responseChan := make(chan *types.SearchResponse, 1)
	errorChan := make(chan error, 1)

	// 监听搜索API请求
	router.MustAdd(DefaultBaseURL+SearchEndpoint, func(ctx *rod.Hijack) {
		// 对于POST请求，我们替换为我们的查询
		if ctx.Request.Method() == "POST" {
			// 修改请求体
			ctx.Request.SetBody(jsonData)

			// 修改头部
			ctx.Request.Req().Header.Set("Content-Type", "application/json")
			ctx.Request.Req().Header.Set("Accept", "application/json")
			ctx.Request.Req().Header.Set("Referer", DefaultBaseURL+"/?query="+query)

			// 执行请求并等待响应
			ctx.MustLoadResponse()

			// 解析响应数据
			var searchResp types.SearchResponse
			body := []byte(ctx.Response.Body())
			err := json.Unmarshal(body, &searchResp)
			if err != nil {
				errorChan <- fmt.Errorf("解析API响应失败: %w", err)
				return
			}

			if bs.debug {
				fmt.Println("API响应:", ctx.Response.Body())
			}

			// 发送响应到通道
			responseChan <- &searchResp
		} else {
			// 其他请求正常继续
			ctx.MustLoadResponse()
		}
	})

	// 开始监听
	go router.Run()

	// 导航到搜索页面或直接触发API调用
	if err := page.Navigate(DefaultBaseURL + "/?query=" + url.QueryEscape(query)); err != nil {
		return nil, fmt.Errorf("导航到搜索页面失败: %w", err)
	}

	// 等待页面加载完成
	if err := page.WaitIdle(30 * time.Second); err != nil {
		return nil, fmt.Errorf("等待搜索页面加载完成失败: %w", err)
	}

	// 等待搜索框出现
	_, err = page.Element("input[type='text']")
	if err == nil {
		// 点击搜索按钮
		searchButton, err := page.Element("button.search-button")
		if err == nil {
			if err := searchButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
				if bs.debug {
					fmt.Println("点击搜索按钮失败:", err)
				}
			}

			// 等待搜索结果加载
			time.Sleep(2 * time.Second)
		}
	}

	// 等待响应或超时
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(10 * time.Second):
		// 如果没有收到API响应，尝试从HTML提取数据
		if bs.debug {
			fmt.Println("API响应超时，尝试从页面提取数据")
		}

		exploits := bs.extractResultsFromPage(page)
		if exploits == nil {
			return nil, fmt.Errorf("获取搜索结果失败: 超时且无法从页面提取数据")
		}

		return exploits, nil
	}
}

// extractResultsFromPage 从页面中提取搜索结果
func (bs *BrowserSearcher) extractResultsFromPage(page *rod.Page) *types.SearchResponse {
	// 此方法在新的实现中不再需要使用，但保留作为备份
	var exploits []types.Exploit

	// 查找所有结果项
	resultItems, err := page.Elements(".exploit-card")
	if err != nil {
		fmt.Println("找不到结果项:", err)
		return nil
	}

	// 提取每个结果的数据
	for _, item := range resultItems {
		var exploit types.Exploit

		// 提取标题
		titleElem, err := item.Element(".exploit-title a")
		if err == nil {
			exploit.Title, _ = titleElem.Text()
			href, err := titleElem.Attribute("href")
			if err == nil && href != nil {
				exploit.Href = *href
			}
		}

		// 提取评分
		scoreElem, err := item.Element(".exploit-score")
		if err == nil {
			score, _ := scoreElem.Text()
			fmt.Sscanf(score, "%f", &exploit.Score)
		}

		// 提取ID
		idElem, err := item.Element(".exploit-id")
		if err == nil {
			exploit.ID, _ = idElem.Text()
		}

		// 提取类型
		typeElem, err := item.Element(".exploit-type")
		if err == nil {
			exploit.Type, _ = typeElem.Text()
		}

		// 提取发布日期
		dateElem, err := item.Element(".exploit-date")
		if err == nil {
			exploit.Published, _ = dateElem.Text()
		}

		if exploit.Title != "" {
			exploits = append(exploits, exploit)
		}
	}

	return &types.SearchResponse{
		Exploits:      exploits,
		ExploitsTotal: len(exploits),
	}
}
