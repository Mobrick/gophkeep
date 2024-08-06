package cardssmigrations

import "embed"

//go:embed *.sql
var EmbedCards embed.FS
