package html

import (
	"embed"
)

// Ref. https://go-review.googlesource.com/c/go/+/359413
//go:embed css images javascript sys usr favicon.ico index.html manifest.json templates translations
var HTML embed.FS
