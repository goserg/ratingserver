JET_OUTPUT="gen"
JET_BOT_OUTPUT="bot/gen"
SQLITE_RATINGS_FILE_LOCATION="rating.sqlite"
SQLITE_BOT_FILE_LOCATION="bot.sqlite"
SERVER_BIN="./bin/server"

gen-jet: build-tools-jet
	$(JET_TOOL) -source=sqlite -dsn=${SQLITE_RATINGS_FILE_LOCATION} -path=${JET_OUTPUT}
	$(JET_TOOL) -source=sqlite -dsn=${SQLITE_BOT_FILE_LOCATION} -path=${JET_BOT_OUTPUT}

build: gen-jet
	go build -o ${SERVER_BIN} cmd/main.go

run: build
	${SERVER_BIN}

# TOOLS

TOOLS_DIR = tools/
TOOLS_MODFILE = $(TOOLS_DIR)go.mod
TOOLS_BIN_DIR = $(TOOLS_DIR)bin/
LINT_TOOL = $(TOOLS_BIN_DIR)golangci-lint
JET_TOOL = $(TOOLS_BIN_DIR)jet

build-tools-lint:
	go build -modfile $(TOOLS_MODFILE) -o $(LINT_TOOL) github.com/golangci/golangci-lint/cmd/golangci-lint

lint: build-tools-lint
	$(LINT_TOOL) run ./...

build-tools-jet:
	go build -modfile $(TOOLS_MODFILE) -o $(JET_TOOL) github.com/go-jet/jet/v2/cmd/jet