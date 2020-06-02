package httppub

import (
	"bytes"
	"strconv"
)

const (
	ContentTypePlainText = "text/plain"
	ContentTypeHTML      = "text/html"
)

// Returns the best available Content-Type to respond to the given Accept header value.
// Our offers are given in order of our preference. Defaults to our first offer.
// Note: x/* matching x/y and x/z isn't implemented, as we don't currently use it.
func Negotiate(accept string, offers ...string) string {
	if len(offers) == 0 {
		return "" // idk what to tell you
	}

	bestOffer := offers[0] // Default to our first offer.
	bestWeight := -1.0     // Default to a very low weight.

	for _, segment := range bytes.Split([]byte(accept), []byte{','}) {
		// Match their offer to ours, accept the one with the highest weight.
		theirs, weight := parseAcceptValue(segment)
		for _, ours := range offers {
			if bytes.Compare(theirs, []byte(ours)) == 0 && weight > bestWeight {
				bestOffer = ours
				bestWeight = weight
			}
		}
	}

	return bestOffer
}

func parseAcceptValue(value []byte) ([]byte, float64) {
	weight := 1.0 // Default weight unless otherwise noted.
	if len(value) == 0 {
		return nil, weight
	}

	// "text/plain; q=0.9" = ["text/plain", "q=0.9"].
	segments := bytes.Split(value, []byte{';'})
	contentType := bytes.TrimFunc(segments[0], func(r rune) bool { return r == ' ' })

	// Everything bar the first index is a parameter.
	for _, param := range segments[1:] {
		// We've got a parameter, it should be in key=value form.
		kvIdx := bytes.IndexByte(param, '=')
		if kvIdx == -1 {
			continue // Not in key=value form.
		}

		// The only parameter we care about is "q".
		key := bytes.TrimFunc(param[:kvIdx], func(r rune) bool { return r == ' ' })
		if len(key) != 1 || key[0] != 'q' {
			continue // Not on the q continuum.
		}

		value := bytes.TrimFunc(param[kvIdx+1:], func(r rune) bool { return r == ' ' })
		w, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			continue // Not a valid float.
		}
		weight = w
	}

	return contentType, weight
}
