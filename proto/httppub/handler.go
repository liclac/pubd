package httppub

import (
	"net/http"
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
	"go.uber.org/zap"

	"github.com/liclac/pubd"
)

type ErrorFn func(err error)

// ErrorFn that logs errors with a warning severity.
func LogErrors(L *zap.Logger) ErrorFn {
	return func(err error) { L.Warn("Error", zap.Error(err)) }
}

// Returns an HTTP handler that serves from a filesystem.
func Handler(fs billy.Filesystem, idxFn Indexer, errCb ErrorFn) http.Handler {
	httpFS := FileSystem(fs) // TODO: Just use a billy.Filesystem.
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := handle(rw, req, httpFS, idxFn); err != nil {
			if errCb != nil {
				errCb(err)
			}
			Error(rw, req, err)
		}
	})
}

// Helper for Handler(), because returning errors is easier.
func handle(rw http.ResponseWriter, req *http.Request, fs http.FileSystem, idxFn Indexer) error {
	if req.Method != http.MethodGet {
		return ErrMethodNotAllowed
	}

	fpath := path.Clean(req.URL.Path)
	if fpath == "" {
		fpath = "/"
	}
	req.URL.Path = fpath

	f, err := fs.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Make sure directory requests end in "/".
		if u := req.URL.Path; u[len(u)-1] != '/' {
			localRedirect(rw, req, path.Base(u)+"/")
			return nil
		}

		// If we have an indexer, render an index.
		if idxFn != nil {
			infos, err := f.Readdir(-1)
			if err != nil {
				return err
			}
			pubd.SortFileInfos(infos)
			if err := idxFn(rw, req, info, infos); err != nil {
				return err
			}
			return nil
		}

		// Else return a 404 Not Found if indexing is not enabled.
		return os.ErrNotExist
	} else {
		// If it's not a directory, it shouldn't end in a slash.
		if u := req.URL.Path; u[len(u)-1] == '/' {
			localRedirect(rw, req, "../"+path.Base(u))
			return nil
		}

		// ServeContent takes care of the rest.
		http.ServeContent(rw, req, info.Name(), info.ModTime(), f)
		return nil
	}
}

// Helper copy-pasted from net/http. Redirects without absolutising newPath.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
