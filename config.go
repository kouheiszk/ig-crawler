package crawler

type Config struct {
	Username       string
	UserAgent      string
	MaxConnections int
	After          int32 // Timestamp
}
