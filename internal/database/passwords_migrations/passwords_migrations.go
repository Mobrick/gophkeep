package passwordsmigrations

import "embed"

//go:embed *.sql
var EmbedPasswords embed.FS
