package html

import (
	"embed"
)

// Ref. https://go-review.googlesource.com/c/go/+/359413
//go:embed css images javascript sys templates usr favicon.ico index.html manifest.json
var HTML embed.FS
