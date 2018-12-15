package ua

import (
	"math/rand"
	"time"
)

func RandomUserAgent() string {
	rand.Seed(time.Now().Unix())
	return UserAgents[rand.Intn(len(UserAgents))]
}
