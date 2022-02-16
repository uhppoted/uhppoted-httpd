package html

import (
	"embed"
)

//go:embed manifest.json css/* images/* javascript/*
var STATIC embed.FS

//go:embed index.html sys/* templates/* usr/*
var HTML embed.FS
