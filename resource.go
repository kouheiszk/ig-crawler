package crawler

import "github.com/kouheiszk/ig-crawler/client"

type Resource struct {
	Url       string
	Timestamp int32
	IsVideo   bool
}

func (r Resource) Data() ([]byte, error) {
	return client.Fetch(r.Url)
}
