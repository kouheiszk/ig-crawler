package crawler

import "github.com/kouheiszk/ig-crawler/pkg/ua"

type Config struct {
	Username       string
	UserAgent      string
	MaxConnections int
	After          int32 // Timestamp
}

func NewConfig() *Config {
	return &Config{
		UserAgent: ua.RandomUserAgent(),
	}
}

func (c *Config) Merge(other *Config) *Config {
	mergeConfig(c, other)
	return c
}

func mergeConfig(dst *Config, other *Config) {
	if other.Username != "" {
		dst.Username = other.Username
	}

	if other.UserAgent != "" {
		dst.UserAgent = other.UserAgent
	}

	if other.MaxConnections != 0 {
		dst.MaxConnections = other.MaxConnections
	}

	if other.After != 0 {
		dst.After = other.After
	}
}
