package mlclassifier

// InnerItem represents the structure of each item in the "items" array.
type DataRequestItem struct {
	Text string `json:"text"`
}

// OuterMap represents the structure of the outer map.
type DataRequest struct {
	Items []DataRequestItem `json:"items"`
}

type DataResponseItem struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}
