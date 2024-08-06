package infosmigrations

import "embed"

//go:embed *.sql
var EmbedInfos embed.FS
