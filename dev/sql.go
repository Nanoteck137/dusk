package dev

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	sqlKeywordRe  = regexp.MustCompile(`(?i)\b(SELECT|FROM|WHERE|AND|OR|ORDER BY|GROUP BY|HAVING|LIMIT|OFFSET|JOIN|LEFT JOIN|RIGHT JOIN|INNER JOIN|OUTER JOIN|CROSS JOIN|ON|INSERT INTO|VALUES|UPDATE|SET|DELETE FROM|CREATE TABLE|ALTER TABLE|DROP TABLE|UNION|UNION ALL|EXCEPT|INTERSECT|CASE|WHEN|THEN|ELSE|END|AS|DISTINCT|IN|NOT|BETWEEN|LIKE|IS|NULL|EXISTS|ASC|DESC)\b`)
	sqlClauseRe   = regexp.MustCompile(`\b(SELECT DISTINCT|SELECT|FROM|LEFT JOIN|RIGHT JOIN|INNER JOIN|CROSS JOIN|JOIN|WHERE|GROUP BY|HAVING|ORDER BY|LIMIT|OFFSET|UNION ALL|UNION|EXCEPT|INTERSECT)\b`)
	sqlWhitespace = regexp.MustCompile(`\s+`)
	sqlLogicalRe  = regexp.MustCompile(`\b(AND|OR)\b`)
)

func PrettyPrintSQL(query string, args []any) string {
	query = substituteParams(query, args)
	query = formatSQL(query)
	return query
}

func substituteParams(query string, args []any) string {
	if len(args) == 0 {
		return query
	}

	result := query
	for i := len(args) - 1; i >= 0; i-- {
		placeholder := fmt.Sprintf("$%d", i+1)
		value := formatValue(args[i])
		result = strings.ReplaceAll(result, placeholder, value)
	}

	for _, arg := range args {
		value := formatValue(arg)
		result = strings.Replace(result, "?", value, 1)
	}

	return result
}

func formatValue(v any) string {
	if v == nil {
		return "NULL"
	}

	switch val := v.(type) {
	case string:
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		escaped := strings.ReplaceAll(string(val), "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("'%v'", val)
	}
}

const sqlIndent = "    "

type clause struct {
	keyword string
	body    string
}

func formatSQL(query string) string {
	query = strings.TrimSpace(query)

	query = sqlKeywordRe.ReplaceAllStringFunc(query, func(match string) string {
		return strings.ToUpper(match)
	})

	query = sqlWhitespace.ReplaceAllString(query, " ")

	indices := sqlClauseRe.FindAllStringIndex(query, -1)
	if len(indices) == 0 {
		return query
	}

	var clauses []clause
	for i, loc := range indices {
		keyword := strings.ToUpper(query[loc[0]:loc[1]])
		bodyStart := loc[1]
		var bodyEnd int
		if i+1 < len(indices) {
			bodyEnd = indices[i+1][0]
		} else {
			bodyEnd = len(query)
		}
		body := strings.TrimSpace(query[bodyStart:bodyEnd])
		clauses = append(clauses, clause{keyword, body})
	}

	var b strings.Builder
	for i, c := range clauses {
		if i > 0 {
			b.WriteString("\n")
		}

		switch c.keyword {
		case "SELECT", "SELECT DISTINCT":
			b.WriteString(c.keyword)
			cols := splitTopLevel(c.body, ',')
			if len(cols) > 1 {
				for j, col := range cols {
					col = strings.TrimSpace(col)
					if j == 0 {
						b.WriteString(" ")
					} else {
						b.WriteString(",\n" + sqlIndent)
					}
					b.WriteString(col)
				}
			} else {
				b.WriteString(" " + c.body)
			}

		case "FROM":
			b.WriteString("FROM " + c.body)

		case "JOIN", "LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "CROSS JOIN":
			b.WriteString(c.keyword + " " + c.body)

		case "WHERE":
			b.WriteString("WHERE")
			conditions := splitLogical(c.body)
			if len(conditions) > 1 {
				for j, cond := range conditions {
					cond = strings.TrimSpace(cond)
					if j == 0 {
						b.WriteString(" " + cond)
					} else {
						b.WriteString("\n" + sqlIndent + cond)
					}
				}
			} else {
				b.WriteString(" " + c.body)
			}

		case "GROUP BY":
			b.WriteString("GROUP BY " + formatCommaList(c.body))

		case "ORDER BY":
			b.WriteString("ORDER BY " + formatCommaList(c.body))

		case "HAVING":
			b.WriteString("HAVING " + c.body)

		case "LIMIT", "OFFSET":
			b.WriteString(c.keyword + " " + c.body)

		case "UNION", "UNION ALL", "EXCEPT", "INTERSECT":
			b.WriteString(c.keyword + "\n")

		default:
			b.WriteString(c.keyword + " " + c.body)
		}
	}

	return b.String()
}

func formatCommaList(s string) string {
	parts := splitTopLevel(s, ',')
	if len(parts) <= 1 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if i == 0 {
			b.WriteString(p)
		} else {
			b.WriteString(",\n" + sqlIndent + p)
		}
	}
	return b.String()
}

func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	start := 0
	inQuote := false
	var quoteChar byte

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inQuote {
			if ch == quoteChar {
				inQuote = false
			}
			continue
		}
		switch ch {
		case '\'', '"':
			inQuote = true
			quoteChar = ch
		case '(':
			depth++
		case ')':
			depth--
		default:
			if ch == sep && depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func splitLogical(s string) []string {
	tokens := sqlLogicalRe.FindAllStringIndex(s, -1)
	if len(tokens) == 0 {
		return []string{s}
	}

	var parts []string
	prev := 0
	for _, loc := range tokens {
		parts = append(parts, strings.TrimSpace(s[prev:loc[0]]))
		parts = append(parts, strings.ToUpper(s[loc[0]:loc[1]]))
		prev = loc[1]
	}
	parts = append(parts, strings.TrimSpace(s[prev:]))

	var result []string
	result = append(result, parts[0])
	for i := 1; i < len(parts); i += 2 {
		conn := parts[i]
		body := ""
		if i+1 < len(parts) {
			body = parts[i+1]
		}
		result = append(result, conn+" "+body)
	}

	return result
}
