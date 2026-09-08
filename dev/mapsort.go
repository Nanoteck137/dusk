// Based on github.com/rogpeppe/go-internal/fmtsort/sort.go
// Original source: https://github.com/rogpeppe/go-internal/blob/master/fmtsort/sort.go
package dev

import (
	"reflect"
	"sort"
)

type sortedMap struct {
	Key   []reflect.Value
	Value []reflect.Value
}

func (o *sortedMap) Len() int           { return len(o.Key) }
func (o *sortedMap) Less(i, j int) bool { return mapCompare(o.Key[i], o.Key[j]) < 0 }
func (o *sortedMap) Swap(i, j int) {
	o.Key[i], o.Key[j] = o.Key[j], o.Key[i]
	o.Value[i], o.Value[j] = o.Value[j], o.Value[i]
}

func mapSort(mapValue reflect.Value) *sortedMap {
	if mapValue.Type().Kind() != reflect.Map {
		return nil
	}

	keys := mapValue.MapKeys()
	values := make([]reflect.Value, len(keys))
	for i, k := range keys {
		values[i] = mapValue.MapIndex(k)
	}

	sorted := &sortedMap{
		Key:   keys,
		Value: values,
	}
	sort.Stable(sorted)
	return sorted
}

func mapCompare(aVal, bVal reflect.Value) int {
	aType, bType := aVal.Type(), bVal.Type()
	if aType != bType {
		return -1
	}
	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		a, b := aVal.Int(), bVal.Int()
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		a, b := aVal.Uint(), bVal.Uint()
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	case reflect.String:
		a, b := aVal.String(), bVal.String()
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	case reflect.Float32, reflect.Float64:
		return floatCompare(aVal.Float(), bVal.Float())
	case reflect.Complex64, reflect.Complex128:
		a, b := aVal.Complex(), bVal.Complex()
		if c := floatCompare(real(a), real(b)); c != 0 {
			return c
		}
		return floatCompare(imag(a), imag(b))
	case reflect.Bool:
		a, b := aVal.Bool(), bVal.Bool()
		switch {
		case a == b:
			return 0
		case a:
			return 1
		default:
			return -1
		}
	case reflect.Ptr:
		a, b := aVal.Pointer(), bVal.Pointer()
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	case reflect.Chan:
		if c, ok := nilCompare(aVal, bVal); ok {
			return c
		}
		ap, bp := aVal.Pointer(), bVal.Pointer()
		switch {
		case ap < bp:
			return -1
		case ap > bp:
			return 1
		default:
			return 0
		}
	case reflect.Struct:
		for i := 0; i < aVal.NumField(); i++ {
			if c := mapCompare(aVal.Field(i), bVal.Field(i)); c != 0 {
				return c
			}
		}
		return 0
	case reflect.Array:
		for i := 0; i < aVal.Len(); i++ {
			if c := mapCompare(aVal.Index(i), bVal.Index(i)); c != 0 {
				return c
			}
		}
		return 0
	case reflect.Interface:
		if c, ok := nilCompare(aVal, bVal); ok {
			return c
		}
		c := mapCompare(reflect.ValueOf(aVal.Elem().Type()), reflect.ValueOf(bVal.Elem().Type()))
		if c != 0 {
			return c
		}
		return mapCompare(aVal.Elem(), bVal.Elem())
	default:
		panic("bad type in compare: " + aType.String())
	}
}

func nilCompare(aVal, bVal reflect.Value) (int, bool) {
	if aVal.IsNil() {
		if bVal.IsNil() {
			return 0, true
		}
		return -1, true
	}
	if bVal.IsNil() {
		return 1, true
	}
	return 0, false
}

func floatCompare(a, b float64) int {
	switch {
	case isNaN(a):
		return -1
	case isNaN(b):
		return 1
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func isNaN(a float64) bool {
	return a != a
}
