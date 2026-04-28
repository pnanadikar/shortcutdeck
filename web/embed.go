package webassets

import "embed"

// FS contains the frontend assets served by the HTTP server.
//
//go:embed index.html app.css app.js
var FS embed.FS
