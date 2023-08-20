.PHONY: all
all: bin/bento

go.mod go.sum:
	go mod tidy

bin/bento: main.go go.mod
	go build -ldflags "-X github.com/catatsuy/bento/cli.Version=`git rev-list HEAD -n1`" -o bin/bento main.go

.PHONY: test
test:
	go test -cover -count 1 ./...
