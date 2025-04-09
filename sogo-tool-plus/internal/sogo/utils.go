package sogo

import (
	"strconv"
	"strings"
)

func (s *DBConfig) normalize(query string) string {
	if s.Driver != "postgres" {
		return query
	}

	var builder strings.Builder
	builder.Grow(len(query)) // Pre-allocate approximate size

	paramIndex := 1
	inSingleQuotes := false
	inDoubleQuotes := false
	escaped := false

	for _, r := range query {
		// Handle the previous character being an escape character
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}

		// Check for escape character
		if r == '\\' {
			builder.WriteRune(r)
			escaped = true
			continue
		}

		// Toggle state if entering/leaving single quotes
		if r == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			builder.WriteRune(r)
			continue
		}

		// Toggle state if entering/leaving double quotes
		if r == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			builder.WriteRune(r)
			continue
		}

		// If we are not inside quotes and find a placeholder
		if r == '?' && !inSingleQuotes && !inDoubleQuotes {
			builder.WriteByte('$')
			builder.WriteString(strconv.Itoa(paramIndex))
			paramIndex++
		} else {
			// Otherwise, just append the character
			builder.WriteRune(r)
		}
	}

	return builder.String()
}
