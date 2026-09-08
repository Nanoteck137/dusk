package dev

import (
	"testing"
)

func TestPrettyPrintSQL(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []any
		expected string
	}{
		{
			name:     "simple select",
			query:    "select * from users where id = ?",
			args:     []any{1},
			expected: "SELECT *\nFROM users\nWHERE id = 1",
		},
		{
			name:     "multiple params",
			query:    "select * from users where name = ? and age > ?",
			args:     []any{"John", 25},
			expected: "SELECT *\nFROM users\nWHERE name = 'John'\n    AND age > 25",
		},
		{
			name:     "dollar placeholders",
			query:    "select * from users where id = $1 and name = $2",
			args:     []any{42, "Alice"},
			expected: "SELECT *\nFROM users\nWHERE id = 42\n    AND name = 'Alice'",
		},
		{
			name:     "null value",
			query:    "select * from users where deleted_at is ?",
			args:     []any{nil},
			expected: "SELECT *\nFROM users\nWHERE deleted_at IS NULL",
		},
		{
			name:     "boolean value",
			query:    "select * from users where active = ?",
			args:     []any{true},
			expected: "SELECT *\nFROM users\nWHERE active = TRUE",
		},
		{
			name:     "string with quotes",
			query:    "select * from users where name = ?",
			args:     []any{"O'Brien"},
			expected: "SELECT *\nFROM users\nWHERE name = 'O''Brien'",
		},
		{
			name:     "no params",
			query:    "select * from users",
			args:     nil,
			expected: "SELECT *\nFROM users",
		},
		{
			name:     "join query",
			query:    "select u.id, p.title from users u join posts p on u.id = p.user_id where u.active = ?",
			args:     []any{true},
			expected: "SELECT u.id,\n    p.title\nFROM users u\nJOIN posts p ON u.id = p.user_id\nWHERE u.active = TRUE",
		},
		{
			name:     "multiple columns",
			query:    "select id, name, email, created_at from users",
			args:     nil,
			expected: "SELECT id,\n    name,\n    email,\n    created_at\nFROM users",
		},
		{
			name:     "complex query",
			query:    "select u.id, u.name, count(p.id) as post_count from users u left join posts p on u.id = p.user_id where u.active = ? and u.age > ? group by u.id, u.name having count(p.id) > ? order by post_count desc limit ?",
			args:     []any{true, 18, 5, 10},
			expected: "SELECT u.id,\n    u.name,\n    count(p.id) AS post_count\nFROM users u\nLEFT JOIN posts p ON u.id = p.user_id\nWHERE u.active = TRUE\n    AND u.age > 18\nGROUP BY u.id,\n    u.name\nHAVING count(p.id) > 5\nORDER BY post_count DESC\nLIMIT 10",
		},
		{
			name:     "or conditions",
			query:    "select * from users where status = ? or role = ? or age > ?",
			args:     []any{"active", "admin", 21},
			expected: "SELECT *\nFROM users\nWHERE status = 'active'\n    OR role = 'admin'\n    OR age > 21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrettyPrintSQL(tt.query, tt.args)
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}
