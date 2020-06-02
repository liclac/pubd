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
		contentType := Negotiate(req.Header.Get("Accept"), ContentTypePlainText, ContentTypeHTML)
		rw.Header().Set("Content-Type", contentType+"; charset=utf-8")

		if contentType == ContentTypeHTML {
			fmt.Fprintf(rw, "<pre>\n")
		}

		for _, d := range infos {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
			u := url.URL{Path: name}

			if contentType == ContentTypeHTML {
				fmt.Fprintf(rw, "<a href=\"%s\">%s</a>\n", u.String(), html.EscapeString(name))
			} else {
				fmt.Fprintln(rw, name)
			}
		}

		if contentType == ContentTypeHTML {
			fmt.Fprintf(rw, "</pre>")
		}
		return nil
	}
}
