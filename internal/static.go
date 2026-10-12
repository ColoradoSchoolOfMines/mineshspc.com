package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"strings"

	"github.com/ColoradoSchoolOfMines/mineshspc.com/website"
)

// The version changes whenever an embedded asset changes, including CSS font files.
var staticVersion = func() string {
	hash := sha256.New()
	err := fs.WalkDir(website.StaticFS, "static", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := website.StaticFS.ReadFile(name)
		if err != nil {
			return err
		}
		hash.Write([]byte(name))
		hash.Write(data)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(hash.Sum(nil))[:16]
}()

func staticURL(name string) string {
	return "/static/v" + staticVersion + "/" + name
}

func staticHandler() http.Handler {
	files := http.FileServer(http.FS(website.StaticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionedPrefix := "/static/v" + staticVersion + "/"
		if after, ok := strings.CutPrefix(r.URL.Path, versionedPrefix); ok {
			name := after
			if _, err := fs.Stat(website.StaticFS, "static/"+name); err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			r = r.Clone(r.Context())
			r.URL.Path = "/static/" + name
		} else if strings.HasPrefix(r.URL.Path, "/static/v") {
			http.NotFound(w, r)
			return
		} else {
			// Keep old links usable, but let clients notice deployments quickly.
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		files.ServeHTTP(w, r)
	})
}
