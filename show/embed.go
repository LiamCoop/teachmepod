// Package show embeds the show bible: personas, seed data, and (later) the
// style guide, slop rules and lexicon. These files are versioned in git.
package show

import "embed"

//go:embed *.yaml
var FS embed.FS
