package pubd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystemExclude(t *testing.T) {
	testdata := map[string]map[string]bool{
		"index.html": map[string]bool{
			"index.txt":             true,
			"index.html":            false,
			"about.html":            true,
			"about.html/wat":        true,
			"subdir/index.html":     false,
			"subdir/index.html/wat": false,
		},
		"*.html": map[string]bool{
			"index.txt":             true,
			"index.html":            false,
			"about.html/wat":        false,
			"subdir/index.html":     false,
			"subdir/index.html/wat": false,
		},
	}
	for exclusion, pathdata := range testdata {
		t.Run(`"`+exclusion+`"`, func(t *testing.T) {
			for path, allowed := range pathdata {
				t.Run(`"`+path+`"`, func(t *testing.T) {
					filter, err := FileSystemConfig{
						Exclude: []string{exclusion},
					}.filter()
					assert.NoError(t, err)
					isAllowed := filter.IsAllowed(path)
					assert.Equal(t, allowed, isAllowed)
				})
			}
		})
	}
}
