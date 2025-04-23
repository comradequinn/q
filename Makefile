.PHONY: build test

NAME=q
BIN=./bin

.PHONY: build
build:
	@go build -o ${BIN}/${NAME}

.PHONY: test
test:
	@go test -count=1 ./...
