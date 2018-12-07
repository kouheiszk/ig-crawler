package crawler

import (
	"strconv"
	"strings"
)

type Resource struct {
	Url       string
	Timestamp int
	IsVideo   bool
}

func (r Resource) Name() string {
	return strconv.Itoa(r.Timestamp) + "_" + strings.Split(r.Url, "/")[len(strings.Split(r.Url, "/"))-1]
}

func (r Resource) Data() ([]byte, error) {
	return fetch(r.Url)
}
