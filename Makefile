.PHONY: all
all: bin/bento

go.mod go.sum:
	go mod tidy

bin/bento: main.go go.mod
	go build -o bin/bento main.go

.PHONY: test
test:
	go test -cover -count 1 ./...
