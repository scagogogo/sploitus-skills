package types

// SearchQuery represents the request body for the Sploitus search API
type SearchQuery struct {
	Type   string `json:"type"`
	Sort   string `json:"sort"`
	Query  string `json:"query"`
	Title  bool   `json:"title"`
	Offset int    `json:"offset"`
}

// Exploit represents a single exploit entry returned by the API
type Exploit struct {
	Title     string  `json:"title"`
	Score     float64 `json:"score"`
	Href      string  `json:"href"`
	Type      string  `json:"type"`
	Published string  `json:"published"`
	ID        string  `json:"id"`
	Source    string  `json:"source"`
	Language  string  `json:"language"`
}

// SearchResponse represents the full response from the Sploitus search API
type SearchResponse struct {
	Exploits      []Exploit `json:"exploits"`
	ExploitsTotal int       `json:"exploits_total"`
}

// ExploitDetail represents detailed information for a single exploit
type ExploitDetail struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Published   string  `json:"published"`
	Source      string  `json:"source"`
	Language    string  `json:"language"`
	Score       float64 `json:"score"`
	Type        string  `json:"type"`
	Href        string  `json:"href"`
}
