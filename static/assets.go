package static

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	platform "github.com/influxdata/influxdb/v2"
)

//go:embed build
var assets embed.FS

const (
	// defaultFile is the default file that will be served if no other static
	// asset matches. this is particularly useful for serving content related to a
	// SPA with client-side routing.
	defaultFile = "index.html"

	// embedPrefix is the prefix for files in the embed.FS - essentially it is
	// the the name of the embedded directory.
	embedPrefix = "build"
)

// NewAssetHandler returns an http.Handler to serve files from the provided
// path. If an empty string is provided as the path, the files are served from
// the embedded assets.
func NewAssetHandler(assetsPath string) http.Handler {
	var assetHandler http.Handler

	if assetsPath != "" {
		assetHandler = fileHandler2(os.DirFS(assetsPath), "")
	} else {
		assetHandler = fileHandler2(assets, embedPrefix)
	}

	return mwSetCacheControl(assetHandler)
}

// mwSetCacheControl sets a default cache control header.
func mwSetCacheControl(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=3600")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// fileHandler takes an fs.FS and a dir name and either returns a handler that
// either serves the file at that path, or the default file if a file cannot be
// found at that path. An empty string can be provided for dir if the files are
// not located in a subdirectory.
func fileHandler2(fileOpener fs.FS, dir string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))

		// If the root directory is being specifically requested, serve the index
		// file and set the content type header, since ServeContent will set it as
		// text/plain otherwise.
		if name == "." {
			name = filepath.Join(dir, defaultFile)
			w.Header().Set("Content-Type", "text/html")
		}

		// Try to open the file requested by name. If it doesn't exist, try to
		// open the index file.
		f, err := fileOpener.Open(filepath.Join(dir, name))
		if err != nil {
			if os.IsNotExist(err) {
				f, err = fileOpener.Open(filepath.Join(dir, defaultFile))
			}
			if err != nil {
				// If the index can't be found, the binary must not have been built with
				// assets, so return no content.
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}
			// Like above, the content type needs to be set for the index file.
			// If we got here, the index must have been found.
			w.Header().Set("Content-Type", "text/html")
		}
		defer f.Close()

		i, err := f.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		content, ok := f.(io.ReadSeeker)
		if !ok {
			// Shouldn't ever get an error here, so return an internal error.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// The etag will be set using some readily available information. The asset
		// files should only change if the user is running a new release, or have
		// set a different asset path. The chance of collisions is low, so a weak
		// tag can be used without much risk.
		etag := fmt.Sprintf(`W/"%s-%d-%s"`, i.Name(), i.Size(), platform.GetBuildInfo().Commit)
		w.Header().Set("ETag", etag)

		// ServeContent will automatically set the content-type header for files
		// other than index.html, and will also set the Last-Modified header.
		// ModTime will be time.Time{} for embedded assets, so a Last-Modified
		// header can't be set.
		http.ServeContent(w, r, name, i.ModTime(), content)
	}

	return http.HandlerFunc(fn)
}

// A much simpler implementation which does not set etags is below....

// fsFunc is short-hand for constructing a http.FileSystem
// implementation
type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

// fileHandler takes an fs.FS and a dir name and either returns a handler that
// either serves the file at that path, or the default file if a file cannot be
// found at that path. An empty string can be provided for dir if the files are
// not located in a subdirectory.
func fileHandler(fileOpener fs.FS, dir string) http.Handler {
	fsys := fsFunc(func(name string) (fs.File, error) {
		f, err := fileOpener.Open(filepath.Join(dir, name))

		if os.IsNotExist(err) {
			return fileOpener.Open(filepath.Join(dir, defaultFile))
		}

		return f, err
	})

	return http.FileServer(http.FS(fsys))
}
