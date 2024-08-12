package filesmigrations

import "embed"

//go:embed *.sql
var EmbedFiles embed.FS
