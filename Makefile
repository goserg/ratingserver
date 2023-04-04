JET_OUTPUT="gen"
SQLITE_FILE_LOCATION="rating.sqlite"
SERVER_BIN="./bin/server"

gen-jet:
	jet -source=sqlite -dsn=${SQLITE_FILE_LOCATION} -path=${JET_OUTPUT}

build:
	go build -o ${SERVER_BIN} cmd/main.go

run: build
	${SERVER_BIN}