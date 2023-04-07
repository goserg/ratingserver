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