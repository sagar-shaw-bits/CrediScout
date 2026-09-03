package web

// Package web embeds the static UI assets into the binary so the server ships
// as a single self-contained artifact (no extra files to mount in Docker/GKE).

import "embed"

//go:embed index.html app.js style.css
var Assets embed.FS
