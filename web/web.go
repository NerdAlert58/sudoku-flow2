// Package web holds the embedded UI assets (contract C5).
package web

import "embed"

//go:embed index.html app.css app.js
var FS embed.FS
