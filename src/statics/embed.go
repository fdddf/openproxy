// Package statics embeds prompt templates and other static assets that the
// providers read at runtime.
package statics

import (
	_ "embed"
)

//go:embed prompt.md
var Prompt string
