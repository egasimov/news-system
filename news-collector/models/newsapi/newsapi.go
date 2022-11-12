package newsapi

type Article struct {
	Source      SrcInfo `json:"source"`
	Author      string  `json:"author"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Url         string  `json:"url"`
	UrlToImage  string  `json:"urlToImage"`
	PublishedAt string  `json:"publishedAt"`
	Content     string  `json:"content"`
}

type SrcInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ApiResponse struct {
	Status       string    `json:"status"`
	TotalResults int32     `json:"totalResults"`
	Articles     []Article `json:"articles"`
}
