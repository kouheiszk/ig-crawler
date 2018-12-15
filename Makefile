build:
	dep ensure -v
	./scripts/make_useragents.sh
	env GOOS=linux go build -ldflags="-s -w" -o bin/crawler crawler.go config.go resource.go fetch.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock
