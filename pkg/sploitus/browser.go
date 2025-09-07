package sploitus

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
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
		Headless(!debug). // 调试模式下显示浏览器窗口
		Devtools(false).  // 关闭DevTools
		UserDataDir("").  // 不使用用户数据目录
		NoSandbox(true).  // 禁用沙箱（在某些环境中可能需要）
		Proxy("")         // 不使用代理

	url := l.MustLaunch()

	// 创建浏览器
	browser := rod.New().
		ControlURL(url).
		Trace(debug)

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

	// 设置自定义Headers
	page.MustEvalOnNewDocument(`() => {
		Object.defineProperty(navigator, 'webdriver', {get: () => false});
	}`)

	// 先访问主页以获取CloudFlare clearance
	fmt.Println("正在访问Sploitus主页...")
	if err := page.Navigate(DefaultBaseURL); err != nil {
		return nil, fmt.Errorf("导航到Sploitus主页失败: %w", err)
	}

	// 等待CloudFlare验证完成
	fmt.Println("等待CloudFlare验证...")
	time.Sleep(5 * time.Second)
	
	// 等待页面加载完成
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("等待页面加载完成失败: %w", err)
	}

	// 现在尝试进行搜索
	fmt.Printf("正在搜索: %s\n", query)
	
	// 查找搜索框并输入查询
	searchBox := page.MustElement("input[type='text']")
	searchBox.MustInput(query)
	time.Sleep(1 * time.Second)
	
	// 按Enter键触发搜索
	searchBox.MustType(input.Enter)
	fmt.Println("已触发搜索，等待结果...")
	
	// 等待搜索结果加载
	time.Sleep(5 * time.Second)
	
	// 直接从页面提取结果
	fmt.Println("正在从页面提取搜索结果...")
	exploits := bs.extractResultsFromPage(page)
	if exploits != nil && len(exploits.Exploits) > 0 {
		fmt.Printf("成功提取到 %d 条结果\n", len(exploits.Exploits))
		return exploits, nil
	}
	
	return nil, fmt.Errorf("获取搜索结果失败: 无法从页面提取数据")
}

// extractResultsFromPage 从页面中提取搜索结果
func (bs *BrowserSearcher) extractResultsFromPage(page *rod.Page) *types.SearchResponse {
	var exploits []types.Exploit
	
	// 尝试多种可能的选择器
	selectors := []string{
		"a[href*='/exploit?id=']",  // Sploitus的漏洞链接
		".text-default",  // 可能的结果项
		"div[class*='card']",
		"div[class*='result']",
		"article",
	}
	
	var resultItems []*rod.Element
	var err error
	
	// 尝试使用最可能的选择器
	for _, selector := range selectors {
		resultItems, err = page.Elements(selector)
		if err == nil && len(resultItems) > 0 {
			if bs.debug {
				fmt.Printf("使用选择器 %s 找到 %d 个元素\n", selector, len(resultItems))
			}
			// 如果是链接选择器，直接处理
			if selector == "a[href*='/exploit?id=']" {
				for _, item := range resultItems {
					var exploit types.Exploit
					
					// 提取标题
					title, _ := item.Text()
					if title != "" {
						exploit.Title = title
					}
					
					// 提取链接
					href, err := item.Attribute("href")
					if err == nil && href != nil {
						exploit.Href = DefaultBaseURL + *href
						// 从URL中提取ID
						if strings.Contains(*href, "id=") {
							parts := strings.Split(*href, "id=")
							if len(parts) > 1 {
								exploit.ID = parts[1]
							}
						}
					}
					
					if exploit.Title != "" {
						exploits = append(exploits, exploit)
					}
				}
				
				if len(exploits) > 0 {
					return &types.SearchResponse{
						Exploits:      exploits,
						ExploitsTotal: len(exploits),
					}
				}
			}
			break
		}
	}
	
	if len(resultItems) == 0 {
		if bs.debug {
			fmt.Println("没有找到任何结果项")
			// 尝试获取页面HTML以便调试
			html, _ := page.HTML()
			if len(html) > 500 {
				fmt.Printf("页面HTML片段: %s...\n", html[:500])
			}
		}
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
