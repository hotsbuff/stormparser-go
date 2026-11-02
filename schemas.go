package stormparser

import "embed"

//go:embed schemas/protocol/builds/*.yaml
var SchemaFiles embed.FS
