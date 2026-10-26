// Package migrations embeds the SQL schema migrations run by `podcastd migrate`.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
