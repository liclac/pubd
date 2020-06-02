package httppub

import (
	"net/http"
	"strings"
)

// Ensures that the prefix for WithPrefix has a leading and trailing '/'.
func CleanPrefix(prefix string) string {
	runes := []rune(prefix)
	if len(runes) == 0 {
		return prefix
	}
	// Trim duplicate leading slash(es).
	if runes[0] == '/' {
		for len(runes) > 1 && runes[1] == '/' {
			runes = runes[1:]
		}
	}
	// Trim trailing slash(es).
	for len(runes) > 0 && runes[len(runes)-1] == '/' {
		runes = runes[:len(runes)-1]
	}
	// Were there nothing but slashes?
	if len(runes) == 0 {
		return ""
	}
	// Build a properly formatted string.
	var b strings.Builder
	if runes[0] != '/' {
		b.Grow(len(runes) + 1)
		b.WriteRune('/')
	} else {
		b.Grow(len(runes))
	}
	for _, r := range runes {
		b.WriteRune(r)
	}
	return b.String()
}

// Serve from a subdirectory, rather than the root. Paths outside prefix 404.
func WithPrefix(prefix string, next http.Handler) http.Handler {
	return http.StripPrefix(CleanPrefix(prefix), next)
}
