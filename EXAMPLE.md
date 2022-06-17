# Complete Guide to Use Squirrel.

### WHERE
```go
import sq "github.com/Masterminds/squirrel"

// Single condition.
query, args, err := sq.Select("*").From("users").Where("id = ?", 1).ToSql()

// Multiple condition.
query, args, err := sq.Select("*").From("users").Where(sq.Eq{
	"age":   15,
	"name": "alex",
}).ToSql()

// Using IN
activeUserIds := []int{1, 2}
query, args, err := sq.Select("*").From("users").Where("id IN (?, ?)", activeUserIds...).ToSql()

// an alternative to using IN
query, args, err := sq.Select("*").From("users").Where(sq.Eq{"id": []int{1, 2, 3, 4, 5}}).ToSql()
```

### INSERT
```go
import sq "github.com/Masterminds/squirrel"

query, args, err := sq.Insert("users").Columns("name", "age").Values("John Doe", 25).ToSql()

// Or using key value map like syntax.
query, args, err := sq.Insert("users").SetMap(sq.Eq{
    "name": "John Doe", 
    "age":  25,
}).ToSql()
```

### UPDATE
```go
import sq "github.com/Masterminds/squirrel"

query, args, err := sq.Update("users").Set("age", 30).Where(sq.Eq{"id": 1}).ToSql()

// Or using key value map like syntax.
query, args, err := sq.Update("users").SetMap(sq.Eq{
    "age": 30
}).Where(sq.Eq{"id": 1}).ToSql()
```

### DELETE
```go
import sq "github.com/Masterminds/squirrel"

query, args, err := sq.Delete("users").Where("id = ?", 1)
```

### SELECT
```go
import sq "github.com/Masterminds/squirrel"

query, args, err := sq.Select("id", "name", "age").From("users").Where("deleted_at IS NULL")
```

### JOIN, LEFT JOIN & RIGHT JOIN
```go
import sq "github.com/Masterminds/squirrel"

query, args, err := sq.
	Select("users.name AS person_name", "class.name AS class_name").
	From("users").
	Join("class ON class.id = users.class_id").
	Where("users.id = ?", 1).
	ToSql()

query, args, err := sq.
	Select("users.name AS person_name", "class.name AS class_name").
	From("users").
	LeftJoin("class ON class.id = users.class_id").
	Where("users.id = ?", 1).
	ToSql()

query, args, err := sq.
	Select("users.name AS person_name", "class.name AS class_name").
	From("users").
	RightJoin("class ON class.id = users.class_id").
	Where("users.id = ?", 1).
	ToSql()
```