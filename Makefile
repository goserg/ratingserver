JET_OUTPUT="gen"
JET_BOT_OUTPUT="bot/gen"
JET_AUTH_OUTPUT="auth/gen"
SQLITE_RATINGS_FILE_LOCATION="rating.sqlite"
SQLITE_BOT_FILE_LOCATION="bot.sqlite"
SQLITE_AUTH_FILE_LOCATION="auth.sqlite"
SERVER_BIN="./bin/server"

gen-jet: build-tools-jet
	$(JET_TOOL) -source=sqlite -dsn=${SQLITE_RATINGS_FILE_LOCATION} -path=${JET_OUTPUT}
	$(JET_TOOL) -source=sqlite -dsn=${SQLITE_AUTH_FILE_LOCATION} -path=${JET_AUTH_OUTPUT}
	$(JET_TOOL) -source=sqlite -dsn=${SQLITE_BOT_FILE_LOCATION} -path=${JET_BOT_OUTPUT}
	$(JET_TOOL) -source=postgres -dsn="postgres://postgres:postgres@localhost:5431/auth?sslmode=disable" -path="gen"

build: gen-jet
	go build -o ${SERVER_BIN} cmd/main.go

run: build atlas-apply
	${SERVER_BIN}

test:
	go list ./... | grep -v tests | xargs go test

TEST_SERVER_CONFIG_PATH = test_configs/server.toml
TEST_BOT_CONFIG_PATH = test_configs/bot.toml

auto-test:
	cd tests && go test -v -server-config $(TEST_SERVER_CONFIG_PATH) -bot-config $(TEST_BOT_CONFIG_PATH) ./...

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
	CGO_ENABLED=1 go build -modfile $(TOOLS_MODFILE) -o $(JET_TOOL) github.com/go-jet/jet/v2/cmd/jet

ATLAS_TOOL = $(TOOLS_BIN_DIR)atlas

build-tools-atlas:
	go build -modfile $(TOOLS_MODFILE) -o $(ATLAS_TOOL) ariga.io/atlas/cmd/atlas

atlas-inspect: build-tools-atlas
	$(ATLAS_TOOL) schema inspect -u "postgres://postgres:postgres@localhost:5431/auth?sslmode=disable" > auth/migrations/init.hcl

atlas-apply: build-tools-atlas
	$(ATLAS_TOOL) schema apply --auto-approve -u "postgres://postgres:postgres@localhost:5431/auth?sslmode=disable" --to file://auth/migrations/init.hcl