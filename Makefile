.PHONY: deps test clean

deps:
	rm -f go.mod go.sum
	go mod init github.com/amaumene/snowfinder_scraper
	go mod edit -replace github.com/amaumene/snowfinder_common=../snowfinder_common_go
	go mod tidy

test:
	go test ./...

clean:
	go clean -cache -testcache
