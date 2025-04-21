.PHONY: build example

NAME=q
BIN=./bin
PROMPT_1="The time is '$$(date)'. What do you know about me?"
PROMPT_2="How many planets in the solar system?"
PROMPT_3="Which is the largest?"

build:
	@go build -o ${BIN}/${NAME}

test:
	@go test -count=1 ./...

example: build
	@${BIN}/${NAME} -n ${PROMPT_1}
	@${BIN}/${NAME} ${PROMPT_2}
	@${BIN}/${NAME} ${PROMPT_3}

example-structured: build
	@${BIN}/${NAME} -n --grounding=false --script --schema='{"type":"object","properties":{"response":{"type":"string"}}}' "pick a colour of the rainbow"

config: build
	@${BIN}/${NAME} --config

sessions: build
	@${BIN}/${NAME} --list

sessions-restore: build
	@${BIN}/${NAME} --restore 1

sessions-delete: build
	@${BIN}/${NAME} --delete 1

sessions-delete-all: build
	@${BIN}/${NAME} --delete-all