build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/crawler crawler.go config.go resource.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock

