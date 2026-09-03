// Package web embeds the built admin UI so the server binary is self-contained.
//
// The Vite build writes into dist/ (see ui/vite.config.js). When the UI has not
// been built the directory holds only .gitkeep, and Available reports false so
// the server can serve a placeholder instead of 404ing every page.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

// Assets returns the built UI filesystem rooted at dist/.
func Assets() fs.FS {
	sub, err := fs.Sub(embedded, "dist")
	if err != nil {
		// dist/ is embedded above, so this cannot happen at runtime.
		panic(err)
	}
	return sub
}

// Available reports whether a UI build is actually embedded.
func Available() bool {
	_, err := fs.Stat(Assets(), "index.html")
	return err == nil
}

// Index returns the SPA entrypoint that unmatched routes fall back to.
func Index() ([]byte, error) {
	return fs.ReadFile(Assets(), "index.html")
}
