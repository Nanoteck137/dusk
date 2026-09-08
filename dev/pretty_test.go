package dev

import (
	"testing"
)

func TestPrintln(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "simple struct",
			value: struct{ Name string }{Name: "test"},
		},
		{
			name:  "nested struct",
			value: struct{ Inner struct{ Value int } }{Inner: struct{ Value int }{Value: 42}},
		},
		{
			name:  "slice",
			value: []int{1, 2, 3, 4, 5},
		},
		{
			name:  "map",
			value: map[string]int{"a": 1, "b": 2},
		},
		{
			name:  "pointer",
			value: &struct{ Name string }{Name: "ptr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sprint(tt.value)
			if result == "" {
				t.Errorf("expected non-empty output")
			}
		})
	}
}

func TestPrintlnOutput(t *testing.T) {
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

	output := Sprint(user)

	expected := `dev.User{
    ID: 1,
    Name: "Alice",
    Address: dev.Address{
        City: "NYC",
        Zip: 10001,
    },
}`

	if output != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, output)
	}
}
