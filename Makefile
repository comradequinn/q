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
	@${BIN}/${NAME} -n --debug "in one sentence, what is the weather like in london tomorrow?" 2> ${BIN}/debug.log
	@${BIN}/${NAME} --flash --debug "in one sentence, what about the day after?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -n --debug -f main.go "in one sentence, summarise this file" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} --debug -f main.go --stats "is it well written?" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -n --debug --schema="colour:string" "pick a colour of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} -n --debug --schema="[]colour:string" "list all colours of the rainbow" 2>> ${BIN}/debug.log
	@${BIN}/${NAME} --list
	@${BIN}/${NAME} --delete-all