package crawler_test

import (
	"fmt"
	"github.com/kouheiszk/ig-crawler"
)

func ExampleCrawler_FetchProfileImage() {
	configs := []crawler.Config{
		{
			Username:       "kouheiszk",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
		{
			Username:       "____invalid____",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
		{
			Username:       "__invalid__",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
	}

	for _, config := range configs {
		url, err := crawler.FetchProfileImage(&config)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(url)
	}

	// Output:
	// https://scontent-nrt1-1.cdninstagram.com/vp/e5bfc3fe9428cd04162c9db6953e78fe/5CA3D19E/t51.2885-19/s320x320/15801844_368469256879175_3717355638490136576_a.jpg
	// not found "https://www.instagram.com/____invalid____/"
	// "__invalid__" is private account
}

func ExampleCrawler_FetchResources() {
	configs := []crawler.Config{
		{
			Username:       "kouheiszk",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
		{
			Username:       "kouheiszk",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
			After:          1499073253,
		},
		{
			Username:       "____invalid____",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
		{
			Username:       "__invalid__",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
			MaxConnections: 2,
		},
	}

	for _, config := range configs {
		resources, err := crawler.FetchResources(&config)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(len(resources), resources[0].Url)
	}

	// Output:
	// 18 https://scontent-nrt1-1.cdninstagram.com/vp/33143be8885222673b04ae16709b0c3f/5C8E40B4/t51.2885-15/e35/19625085_409871079414311_3741424003856728064_n.jpg
	// 4 https://scontent-nrt1-1.cdninstagram.com/vp/cc04e8ca762a14223e8b29879ad85a82/5C92BAFB/t51.2885-15/e35/47183741_1947269182245762_213590525911104973_n.jpg
	// not found "https://www.instagram.com/____invalid____/"
	// "__invalid__" is private account
}
