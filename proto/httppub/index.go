package httppub

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
)

type IndexConfig struct {
	// README files are included at the bottom of a directory listing.
	READMEs []string `toml:"readme"`
}

// Interface for producing a directory index.
type Indexer interface {
	Render(http.ResponseWriter, *http.Request, billy.Filesystem, []os.FileInfo) error
}

type simpleIndex struct {
	READMEs map[string]bool
}

// Generate a simple index, very similar to the one used by http.FileServer.
func SimpleIndex(cfg IndexConfig) Indexer {
	idx := simpleIndex{READMEs: make(map[string]bool)}
	for _, filename := range cfg.READMEs {
		idx.READMEs[filename] = true
	}
	return idx
}

func (idx simpleIndex) Render(rw http.ResponseWriter, req *http.Request, fs billy.Filesystem, infos []os.FileInfo) error {
	// Figure out if we're rendering plain text, or in a thin HTML wrapper.
	contentType := Negotiate(req.Header.Get("Accept"), ContentTypePlainText, ContentTypeHTML)
	rw.Header().Set("Content-Type", contentType+"; charset=utf-8")

	if contentType == ContentTypeHTML {
		fmt.Fprintf(rw, "<pre>\n")
	}

	// Print a directory listing.
	foundREADME := ""
	for _, info := range infos {
		name := info.Name()
		switch {
		case info.IsDir():
			name += "/"
		case idx.READMEs[name]:
			foundREADME = name
		}

		if contentType == ContentTypeHTML {
			fmt.Fprintf(rw, "<a href=\"%s\">%s</a>\n",
				(&url.URL{Path: name}).String(), html.EscapeString(name))
		} else {
			fmt.Fprintln(rw, name)
		}
	}

	// If we have a README, tuck that on at the bottom. An error here is non-fatal.
	// TODO: String along a logger through this code, we should still log a warning.
	if foundREADME != "" {
		_ = idx.renderREADME(rw, fs, filepath.Join(req.URL.Path, foundREADME))
	}

	if contentType == ContentTypeHTML {
		fmt.Fprintf(rw, "</pre>")
	}
	return nil
}

func (idx simpleIndex) renderREADME(w io.Writer, fs billy.Filesystem, path string) error {
	f, err := fs.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprint(w, "\n")
	_, _ = io.Copy(w, f)
	return nil
}
