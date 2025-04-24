.PHONY: build test

NAME=q
BIN=./bin

.PHONY: build
build:
	@go build -o ${BIN}/${NAME}

.PHONY: test
test:
	@go test -count=1 ./...

.PHONY: examples
examples: 
	@${BIN}/${NAME} --delete-all
	@${BIN}/${NAME} -n "in one sentence, what is the weather like in london tomorrow?"
	@${BIN}/${NAME} --flash "in one sentence, what about the day after?"
	@${BIN}/${NAME} -n -f main.go "in one sentence, summarise this file"
	@${BIN}/${NAME} -n --schema='colour:string' "pick a colour of the rainbow"
	@${BIN}/${NAME} --list
	@${BIN}/${NAME} --delete-all