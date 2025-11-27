package templates

import "embed"

// FS contains all embedded template files.
//
//go:embed *.tmpl
var FS embed.FS
