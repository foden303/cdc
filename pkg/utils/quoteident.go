package utils

import (
	"strings"
)

// QuoteIdentifierDoubleQuote quotes a SQL identifier with double-quotes.
func QuoteIdentifierDoubleQuote(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return quoteIdentPart(parts[0]) + "." + quoteIdentPart(parts[1])
	}
	return quoteIdentPart(name)
}

// QuoteIdentifierBacktick quotes a SQL identifier with backticks.
func QuoteIdentifierBacktick(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return quoteBacktick(parts[0]) + "." + quoteBacktick(parts[1])
	}
	return quoteBacktick(name)
}

// quoteIdentPart wraps a single identifier in double-quotes.
func quoteIdentPart(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
}

// quoteBacktick wraps a single identifier in backticks.
func quoteBacktick(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}
