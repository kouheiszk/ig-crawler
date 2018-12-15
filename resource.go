package crawler

type Resource struct {
	Url       string `json:"url"`
	Timestamp int32  `json:"timestamp"`
	IsVideo   bool   `json:"is_video"`
}
