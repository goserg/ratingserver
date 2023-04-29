JET_OUTPUT="gen"
JET_BOT_OUTPUT="bot/gen"
SQLITE_RATINGS_FILE_LOCATION="rating.sqlite"
SQLITE_BOT_FILE_LOCATION="bot.sqlite"
SERVER_BIN="./bin/server"

gen-jet:
	jet -source=sqlite -dsn=${SQLITE_RATINGS_FILE_LOCATION} -path=${JET_OUTPUT}
	jet -source=sqlite -dsn=${SQLITE_BOT_FILE_LOCATION} -path=${JET_BOT_OUTPUT}

build:
	go build -o ${SERVER_BIN} cmd/main.go

run: build
	${SERVER_BIN}

# TOOLS

TOOLS_DIR = tools/
TOOLS_MODFILE = $(TOOLS_DIR)go.mod
TOOLS_BIN_DIR = $(TOOLS_DIR)bin/
LINT_TOOL = $(TOOLS_BIN_DIR)golangci-lint

build-tools-lint:
	go build -modfile $(TOOLS_MODFILE) -o $(LINT_TOOL) github.com/golangci/golangci-lint/cmd/golangci-lint

PHONY: lint
lint: build-tools-lint
	$(LINT_TOOL) run ./...