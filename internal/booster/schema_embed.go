package booster

import _ "embed"

// SchemaJSON is the JSON Schema for booster.toml, embedded at build time.
// It is written to .booster/booster.schema.json by booster install so that
// editors with Taplo/Even Better TOML support pick it up automatically.
//
//go:embed schema/booster.schema.json
var SchemaJSON string
