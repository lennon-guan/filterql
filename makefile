
.PHONY: test parse

parse:
	go build -o bin/parse cmd/parse/*.go

test:
	go test -v filter_test.go
