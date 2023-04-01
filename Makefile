JET_OUTPUT="gen"
SQLITE_FILE_LOCATION="rating.sqlite"

gen-jet:
	jet -source=sqlite -dsn=${SQLITE_FILE_LOCATION} -path=${JET_OUTPUT}
