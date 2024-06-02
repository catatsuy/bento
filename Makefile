.PHONY: all
all: bin/bento

go.mod go.sum:
	go mod tidy

bin/bento: main.go go.mod $(wildcard internal/**/*.go)
	go build -ldflags "-X github.com/catatsuy/bento/internal/cli.Version=`git rev-list HEAD -n1`" -o bin/bento main.go

.PHONY: test
test:
	go test -cover -count 1 ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: errcheck
errcheck:
	errcheck ./...

.PHONY: staticcheck
staticcheck:
	staticcheck -checks="all,-ST1000" ./...

.PHONY: clean
clean:
	rm -rf bin/*
