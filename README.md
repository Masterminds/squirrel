# Squirrel - fluent SQL generator for Go

[![GoDoc](https://godoc.org/github.com/lann/squirrel?status.png)](https://godoc.org/github.com/lann/squirrel)
[![Build Status](https://travis-ci.org/lann/squirrel.png?branch=master)](https://travis-ci.org/lann/squirrel)

**Squirrel is not an ORM.**

Squirrel helps you build SQL queries from composable parts:

```go
users := Select("*").From("users")

active := users.Where(Eq{"deleted_at": nil})

sql, args, err := active.ToSql()

sql == "SELECT * FROM users WHERE deleted_at IS NULL"
```

Squirrel can also execute queries directly:

```go
stooges := users.Where(Eq{"username": []string{"moe", "larry", "curly", "shemp"}})
three_stooges := stooges.Limit(3)
rows, err := three_stooges.RunWith(db).Query()

// Behaves like:
rows, err := db.Query("SELECT * FROM users WHERE username IN (?,?,?,?) LIMIT 3",
                      "moe", "larry", "curly", "shemp")
```

Squirrel wants to make your life easier:

```go
// StmtCache caches Prepared Stmts for you
dbCache := NewStmtCache(db)

// StatementBuilder keeps your syntax neat
mydb := StatementBuilder.RunWith(dbCache)
select_users := mydb.Select("*").From("users")
```
