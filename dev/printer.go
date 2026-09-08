// Based on github.com/kr/pretty/formatter.go
// Original source: https://github.com/kr/pretty/blob/master/formatter.go
package dev

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

func Println(a ...any) (n int, err error) {
	return fmt.Println(wrap(a)...)
}

func Sprint(a ...any) string {
	return fmt.Sprint(wrap(a)...)
}

func wrap(a []any) []any {
	w := make([]any, len(a))
	for i, x := range a {
		w[i] = formatter{v: reflect.ValueOf(x), force: true, quote: true}
	}
	return w
}

type formatter struct {
	v     reflect.Value
	force bool
	quote bool
}

func (fo formatter) String() string {
	return fmt.Sprint(fo.v.Interface())
}

func (fo formatter) passThrough(f fmt.State, c rune) {
	var s strings.Builder
	s.WriteString("%")

	for i := range 128 {
		if f.Flag(i) {
			s.WriteString(string(rune(i)))
		}
	}

	if w, ok := f.Width(); ok {
		fmt.Fprintf(&s, "%d", w)
	}

	if p, ok := f.Precision(); ok {
		fmt.Fprintf(&s, ".%d", p)
	}

	s.WriteString(string(c))
	fmt.Fprintf(f, s.String(), fo.v.Interface())
}

func (fo formatter) Format(f fmt.State, c rune) {
	if fo.force || c == 'v' && f.Flag('#') && f.Flag(' ') {
		w := newWriter(f)
		p := &printer{w: w, visited: make(map[visit]int)}
		p.printValue(fo.v, true, fo.quote)
		return
	}
	fo.passThrough(f, c)
}

type writer struct {
	w           io.Writer
	indentLevel int
	indentStr   string
}

func newWriter(w io.Writer) *writer {
	return &writer{w: w, indentStr: "    "}
}

func (w *writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w *writer) writeByte(b byte) {
	w.w.Write([]byte{b})
}

func (w *writer) writeString(s string) {
	io.WriteString(w.w, s)
}

func (w *writer) writeIndent() {
	for i := 0; i < w.indentLevel; i++ {
		io.WriteString(w.w, w.indentStr)
	}
}

func (w *writer) indent() {
	w.indentLevel++
}

func (w *writer) unindent() {
	w.indentLevel--
	if w.indentLevel < 0 {
		w.indentLevel = 0
	}
}

type printer struct {
	w       *writer
	visited map[visit]int
	depth   int
}

func (p *printer) printInline(v reflect.Value, x any, showType bool) {
	if showType {
		p.w.writeString(typeName(v.Type()))
		fmt.Fprintf(p.w, "(%#v)", x)
	} else {
		fmt.Fprintf(p.w, "%#v", x)
	}
}

type visit struct {
	v   uintptr
	typ reflect.Type
}

func (p *printer) catchPanic(v reflect.Value, method string) {
	if r := recover(); r != nil {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			p.w.writeByte('(')
			p.w.writeString(typeName(v.Type()))
			p.w.writeString(")(nil)")
			return
		}

		p.w.writeByte('(')
		p.w.writeString(typeName(v.Type()))
		p.w.writeString(")(PANIC=calling method ")
		p.w.writeString(strconv.Quote(method))
		p.w.writeString(": ")
		fmt.Fprint(p.w, r)
		p.w.writeByte(')')
	}
}

func (p *printer) printValue(v reflect.Value, showType, quote bool) {
	if p.depth > 10 {
		p.w.writeString("!%v(DEPTH EXCEEDED)")
		return
	}

	if v.IsValid() && v.CanInterface() {
		i := v.Interface()
		if goStringer, ok := i.(fmt.GoStringer); ok {
			defer p.catchPanic(v, "GoString")
			p.w.writeString(goStringer.GoString())
			return
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		p.printInline(v, v.Bool(), showType)
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		p.printInline(v, v.Int(), showType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		p.printInline(v, v.Uint(), showType)
	case reflect.Float32, reflect.Float64:
		p.printInline(v, v.Float(), showType)
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprintf(p.w, "%#v", v.Complex())
	case reflect.String:
		p.fmtString(v.String(), quote)
	case reflect.Map:
		p.printMap(v, showType)
	case reflect.Struct:
		p.printStruct(v, showType)
	case reflect.Interface:
		switch e := v.Elem(); {
		case e.Kind() == reflect.Invalid:
			p.w.writeString("nil")
		case e.IsValid():
			p.depth++
			p.printValue(e, showType, true)
			p.depth--
		default:
			p.w.writeString(typeName(v.Type()))
			p.w.writeString("(nil)")
		}
	case reflect.Array, reflect.Slice:
		p.printSlice(v, showType)
	case reflect.Pointer:
		e := v.Elem()
		if !e.IsValid() {
			p.w.writeByte('(')
			p.w.writeString(typeName(v.Type()))
			p.w.writeString(")(nil)")
		} else {
			p.depth++
			p.w.writeByte('&')
			p.printValue(e, true, true)
			p.depth--
		}
	case reflect.Chan:
		x := v.Pointer()
		if showType {
			p.w.writeByte('(')
			p.w.writeString(typeName(v.Type()))
			fmt.Fprintf(p.w, ")(%#v)", x)
		} else {
			fmt.Fprintf(p.w, "%#v", x)
		}
	case reflect.Func:
		p.w.writeString(typeName(v.Type()))
		p.w.writeString(" {...}")
	case reflect.UnsafePointer:
		p.printInline(v, v.Pointer(), showType)
	case reflect.Invalid:
		p.w.writeString("nil")
	}
}

func (p *printer) printMap(v reflect.Value, showType bool) {
	t := v.Type()

	if showType {
		p.w.writeString(typeName(t))
	}

	p.w.writeByte('{')

	if !nonzero(v) {
		p.w.writeByte('}')
		return
	}

	expand := !canInline(t.Elem())
	pp := p

	if expand {
		p.w.writeByte('\n')
		p.w.indent()
	}

	sm := mapSort(v)

	for i := 0; i < v.Len(); i++ {
		k := sm.Key[i]
		mv := sm.Value[i]

		if expand {
			p.w.writeIndent()
		}

		pp.printValue(k, false, true)
		pp.w.writeByte(':')
		pp.w.writeByte(' ')

		showTypeInMap := t.Elem().Kind() == reflect.Interface
		pp.printValue(mv, showTypeInMap, true)

		if expand {
			p.w.writeString(",\n")
		} else if i < v.Len()-1 {
			p.w.writeString(", ")
		}
	}

	if expand {
		p.w.unindent()
	}

	p.w.writeByte('}')
}

func (p *printer) printStruct(v reflect.Value, showType bool) {
	t := v.Type()

	if v.CanAddr() {
		addr := v.UnsafeAddr()
		vis := visit{addr, t}
		if vd, ok := p.visited[vis]; ok && vd < p.depth {
			p.fmtString(typeName(t)+"{(CYCLIC REFERENCE)}", false)
			return
		}
		p.visited[vis] = p.depth
	}

	if showType {
		p.w.writeString(typeName(t))
	}

	p.w.writeByte('{')

	if !nonzero(v) {
		p.w.writeByte('}')
		return
	}

	p.w.writeByte('\n')
	p.w.indent()

	for i := 0; i < v.NumField(); i++ {
		p.w.writeIndent()

		showTypeInStruct := true
		if f := t.Field(i); f.Name != "" {
			p.w.writeString(f.Name)
			p.w.writeByte(':')
			p.w.writeByte(' ')
			showTypeInStruct = labelType(f.Type)
		}

		p.depth++
		p.printValue(getField(v, i), showTypeInStruct, true)
		p.depth--
		p.w.writeString(",\n")
	}

	p.w.unindent()
	p.w.writeIndent()
	p.w.writeByte('}')
}

func (p *printer) printSlice(v reflect.Value, showType bool) {
	t := v.Type()

	if showType {
		p.w.writeString(typeName(t))
	}

	if v.Kind() == reflect.Slice && v.IsNil() {
		if showType {
			p.w.writeString("(nil)")
		} else {
			p.w.writeString("nil")
		}

		return
	}

	p.w.writeByte('{')
	expand := !canInline(t.Elem())
	pp := p
	if expand {
		p.w.writeByte('\n')
		p.w.indent()
	}

	for i := 0; i < v.Len(); i++ {
		if expand {
			p.w.writeIndent()
		}

		showTypeInSlice := t.Elem().Kind() == reflect.Interface
		pp.printValue(v.Index(i), showTypeInSlice, true)
		if expand {
			p.w.writeString(",\n")
		} else if i < v.Len()-1 {
			p.w.writeString(", ")
		}
	}

	if expand {
		p.w.unindent()
	}

	p.w.writeByte('}')
}

func canInline(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map:
		return !canExpand(t.Elem())
	case reflect.Struct:
		return false
	case reflect.Interface:
		return false
	case reflect.Array, reflect.Slice:
		return !canExpand(t.Elem())
	case reflect.Pointer:
		return false
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	}
	return true
}

func canExpand(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Struct,
		reflect.Interface, reflect.Array, reflect.Slice,
		reflect.Pointer:
		return true
	}
	return false
}

func labelType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Struct:
		return true
	}
	return false
}

func (p *printer) fmtString(s string, quote bool) {
	if quote {
		s = strconv.Quote(s)
	}
	p.w.writeString(s)
}

func getField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}

	return val
}

func typeName(t reflect.Type) string {
	return shortenTypeNames(t.String())
}

func shortenTypeNames(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		if s[i] == '[' || s[i] == ',' || s[i] == ' ' ||
			s[i] == '*' || s[i] == ']' {
			result.WriteByte(s[i])
			i++
			continue
		}

		j := i
		for j < len(s) && s[j] != '[' && s[j] != ']' &&
			s[j] != ',' && s[j] != ' ' {
			j++
		}

		result.WriteString(shortenTypeName(s[i:j]))
		i = j
	}

	return result.String()
}

func shortenTypeName(name string) string {
	if strings.HasPrefix(name, "[]") {
		return "[]" + shortenTypeName(name[2:])
	}

	if strings.HasPrefix(name, "*") {
		return "*" + shortenTypeName(name[1:])
	}

	if strings.HasPrefix(name, "map[") {
		rest := name[4:]
		bracketCount := 1
		i := 0

		for i < len(rest) && bracketCount > 0 {
			switch rest[i] {
			case '[':
				bracketCount++
			case ']':
				bracketCount--
			}

			i++
		}

		keyType := rest[:i-1]
		valueType := rest[i:]

		return "map[" + shortenTypeName(keyType) + "]" +
			shortenTypeName(valueType)
	}

	if idx := strings.Index(name, "["); idx >= 0 {
		baseName := name[:idx]
		params := name[idx:]

		if dotIdx := strings.LastIndex(baseName, "."); dotIdx >= 0 {
			pkg := baseName[:dotIdx]
			if strings.Contains(pkg, ".") {
				baseName = baseName[dotIdx+1:]
			}
		}

		return baseName + shortenTypeNames(params)
	}

	if dotIdx := strings.LastIndex(name, "."); dotIdx >= 0 {
		pkg := name[:dotIdx]
		if strings.Contains(pkg, ".") {
			slashIdx := strings.LastIndex(pkg, "/")
			if slashIdx >= 0 {
				return pkg[slashIdx+1:] + "." + name[dotIdx+1:]
			}
		}
	}

	return name
}
