package resources

import "embed"

//go:embed themes/*.json
var Themes embed.FS
