.PHONY: build test

NAME=gen
BIN=./bin

.PHONY: build
build:
	@go build -o ${BIN}/${NAME}

.PHONY: test
test:
	@go test -count=1 ./...

.PHONY: examples
examples: build
	@${BIN}/${NAME} --delete-all
	@${BIN}/${NAME} -n "in one sentence, what is the weather like in london tomorrow?" 2> ${BIN}/debug.log
	@${BIN}/${NAME} --flash "in one sentence, what about the day after?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -n -f main.go "in one sentence, summarise this file" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -f main.go --stats "is it well written?" 2> stats.txt 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -n --schema='colour:string' "pick a colour of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} --list
	@${BIN}/${NAME} --delete-all
	@echo ""
	@cat stats.txt && rm stats.txt
