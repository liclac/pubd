package httppub

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"os"
)

// Function called to produce a directory index.
type Indexer func(rw http.ResponseWriter, req *http.Request, dir os.FileInfo, infos []os.FileInfo) error

// Generate a simple index, very similar to the one used by http.FileServer.
func SimpleIndex() Indexer {
	return func(rw http.ResponseWriter, req *http.Request, dir os.FileInfo, infos []os.FileInfo) error {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(rw, "<pre>\n")
		for _, d := range infos {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
			u := url.URL{Path: name}
			fmt.Fprintf(rw, "<a href=\"%s\">%s</a>\n", u.String(), html.EscapeString(name))
		}

		fmt.Fprintf(rw, "</pre>\n")
		return nil
	}
}
