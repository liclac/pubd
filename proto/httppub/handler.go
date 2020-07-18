package httppub

import (
	"net/http"
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
	"go.uber.org/zap"

	"github.com/liclac/pubd"
)

// Returns an HTTP handler that serves from a filesystem.
func Handler(L *zap.Logger, fs billy.Filesystem, idxFn Indexer) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := handle(L, rw, req, fs, idxFn); err != nil {
			RenderError(rw, req, err)
		}
	})
}

// Helper for Handler(), because returning errors is easier.
func handle(L *zap.Logger, rw http.ResponseWriter, req *http.Request, fs billy.Filesystem, idxFn Indexer) error {
	if req.Method != http.MethodGet {
		return ErrMethodNotAllowed
	}

	info, err := fs.Stat(req.URL.Path)
	if err != nil {
		return err
	}

	isDir := info.IsDir()
	if cpath := redirectForCanon(req.URL.Path, isDir); cpath != "" {
		localRedirect(rw, req, cpath)
		return nil
	}

	if isDir {
		// If we have an indexer, render an index.
		if idxFn != nil {
			infos, err := fs.ReadDir(req.URL.Path)
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
		f, err := fs.Open(req.URL.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		// ServeContent takes care of the rest.
		http.ServeContent(rw, req, info.Name(), info.ModTime(), f)
		return nil
	}
}

// Clean a request path.
func cleanPath(in string) string {
	out := []rune(path.Clean("/" + in))
	// The "/" prefix should ensure that this never happens, but still.
	if len(out) == 0 {
		return "/"
	}
	// If the input path ends in a "/", so shall the output path.
	if in != "" && in[len(in)-1] == '/' && out[len(out)-1] != '/' {
		out = append(out, '/')
	}
	return string(out)
}

// Requests for directories should have paths ending in "/", files should not.
func redirectForCanon(p string, isDir bool) string {
	if isDir {
		if len(p) == 0 {
			return "./"
		} else if p[len(p)-1] != '/' {
			return path.Base(p) + "/"
		}
	} else if len(p) > 0 && p[len(p)-1] == '/' {
		return path.Clean("../" + path.Base(p))
	}
	return ""
}

// Helper copy-pasted from net/http. Redirects without absolutising newPath.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}
