build:
	dep ensure -v
	./scripts/make_useragents.sh
	env GOOS=linux go build -ldflags="-s -w" -o bin/crawler cmd/crawler/main.go

.PHONY: clean
clean:
	rm -rf ./bin ./vendor Gopkg.lock
