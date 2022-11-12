package theguardian

type ApiResponse struct {
	Response Response `json:"response"`
}

type Response struct {
	Status     string   `json:"status"`
	UserTier   string   `json:"userTier"`
	Total      int32    `json:"total"`
	StartIndex int32    `json:"startIndex"`
	PageSize   int32    `json:"pageSize"`
	CurrenPage int32    `json:"currenPage"`
	Pages      int32    `json:"pages"`
	Results    []Result `json:"results"`
}

type Result struct {
	Id                 string `json:"id"`
	Type               string `json:"type"`
	SectionId          string `json:"sectionId"`
	SectionName        string `json:"sectionName"`
	WebPublicationDate string `json:"webPublicationDate"`
	WebTitle           string `json:"webTitle"`
	WebUrl             string `json:"webUrl"`
	ApiUrl             string `json:"apiUrl"`
	IsHosted           bool   `json:"isHosted"`
	PillarId           string `json:"pillarId"`
	PillarName         string `json:"pillarName"`
}
