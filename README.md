# Squirrel - fluent SQL generator for Go

```go
import "github.com/lann/squirrel"
```

[![GoDoc](https://godoc.org/github.com/lann/squirrel?status.png)](https://godoc.org/github.com/lann/squirrel)
[![Build Status](https://travis-ci.org/lann/squirrel.png?branch=master)](https://travis-ci.org/lann/squirrel)

**Squirrel is not an ORM.**

Squirrel helps you build SQL queries from composable parts:

```go
users := squirrel.Select("*").From("users")

active := users.Where(Eq{"deleted_at": nil})

sql, args, err := active.ToSql()

sql == "SELECT * FROM users WHERE deleted_at IS NULL"
```

Squirrel can also execute queries directly:

```go
stooges := users.Where(squirrel.Eq{"username": []string{"moe", "larry", "curly", "shemp"}})
three_stooges := stooges.Limit(3)
rows, err := three_stooges.RunWith(db).Query()

// Behaves like:
rows, err := db.Query("SELECT * FROM users WHERE username IN (?,?,?,?) LIMIT 3",
                      "moe", "larry", "curly", "shemp")
```

Squirrel makes conditional query building a breeze:

```go
if len(q) > 0 {
    users = users.Where("name LIKE ?", fmt.Sprint("%", q, "%"))
}
```

Squirrel wants to make your life easier:

```go
// StmtCache caches Prepared Stmts for you
dbCache := squirrel.NewStmtCache(db)

// StatementBuilder keeps your syntax neat
mydb := squirrel.StatementBuilder.RunWith(dbCache)
select_users := mydb.Select("*").From("users")
```

Squirrel loves PostgreSQL:

```go
psql := squirrel.StatementBuilder.PlaceholderFormat(Dollar)

// You use question marks for placeholders...
sql, _, _ := psql.Select("*").From("elephants").Where("name IN (?,?)", "Dumbo", "Verna")

/// ...squirrel replaces them using PlaceholderFormat.
sql == "SELECT * FROM elephants WHERE name IN ($1,$2)"
```

## License

Builder is released under the
[MIT License](http://www.opensource.org/licenses/MIT).
