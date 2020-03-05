module github.com/Masterminds/squirrel/integration

go 1.12

require (
	github.com/Masterminds/squirrel v1.1.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/lib/pq v1.2.0
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/stretchr/testify v1.4.0
	google.golang.org/appengine v1.6.5 // indirect
)

replace github.com/Masterminds/squirrel => ../
