# dev

Development utilities for debugging. This package provides pretty-printing for Go values and SQL statements.

## Pretty Printing

The `Println` and `Sprint` functions provide pretty-printing for Go values, similar to `github.com/kr/pretty`.

### Examples

**Structs**
```go
type User struct {
    ID   int
    Name string
}
user := User{ID: 1, Name: "Alice"}
dev.Println(user)
```
Output:
```
dev.User{
    ID: 1,
    Name: "Alice",
}
```

**Nested structs**
```go
type Address struct {
    City string
    Zip  int
}
type User struct {
    ID      int
    Name    string
    Address Address
}
user := User{ID: 1, Name: "Alice", Address: Address{City: "NYC", Zip: 10001}}
dev.Println(user)
```
Output:
```
dev.User{
    ID: 1,
    Name: "Alice",
    Address: dev.Address{
        City: "NYC",
        Zip: 10001,
    },
}
```

**Slices**
```go
items := []string{"apple", "banana", "cherry"}
dev.Println(items)
```
Output:
```
[]string{
    "apple",
    "banana",
    "cherry",
}
```

**Maps**
```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
dev.Println(m)
```
Output:
```
map[string]int{
    "a": 1,
    "b": 2,
    "c": 3,
}
```

## SQL Pretty Printing

The `PrettyPrintSQL` function formats SQL statements with proper indentation and parameter substitution.

### Examples

**Simple query with parameters**
```go
sql := "select * from users where name = ? and age > ?"
params := []any{"John", 25}
fmt.Println(dev.PrettyPrintSQL(sql, params))
```
Output:
```sql
SELECT *
FROM users
WHERE name = 'John'
    AND age > 25
```

**Complex query**
```go
sql := "select u.id, u.name, count(p.id) as post_count from users u left join posts p on u.id = p.user_id where u.active = ? and u.age > ? group by u.id, u.name having count(p.id) > ? order by post_count desc limit ?"
params := []any{true, 18, 5, 10}
fmt.Println(dev.PrettyPrintSQL(sql, params))
```
Output:
```sql
SELECT u.id,
    u.name,
    count(p.id) AS post_count
FROM users u
LEFT JOIN posts p ON u.id = p.user_id
WHERE u.active = TRUE
    AND u.age > 18
GROUP BY u.id,
    u.name
HAVING count(p.id) > 5
ORDER BY post_count DESC
LIMIT 10
```

**Dollar placeholders**
```go
sql := "select * from users where id = $1 and name = $2"
params := []any{42, "Alice"}
fmt.Println(dev.PrettyPrintSQL(sql, params))
```
Output:
```sql
SELECT *
FROM users
WHERE id = 42
    AND name = 'Alice'
```

## Credits

The pretty printing functionality is based on:

- [github.com/kr/pretty](https://github.com/kr/pretty) - Pretty printing for Go values
- [github.com/kr/text](https://github.com/kr/text) - Text manipulation utilities (indentWriter)
- [github.com/rogpeppe/go-internal/fmtsort](https://github.com/rogpeppe/go-internal) - Map sorting utilities

These dependencies have been vendored directly into this package to avoid external dependencies.
