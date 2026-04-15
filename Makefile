.PHONY: deps test clean

deps:
	rm -f go.mod go.sum
	go mod init github.com/amaumene/snowfinder_common
	go mod tidy

test:
	go test ./...

clean:
	rm -f go.mod go.sum
	go clean -cache -testcache
