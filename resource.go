package crawler

type Resource struct {
	Url       string
	Timestamp int32
	IsVideo   bool
}

func (r Resource) Data() ([]byte, error) {
	return fetch(r.Url)
}
