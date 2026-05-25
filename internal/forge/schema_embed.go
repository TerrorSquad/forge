package forge

import _ "embed"

// SchemaJSON is the JSON Schema for forge.toml, embedded at build time.
// It is written to .forge/forge.schema.json by forge install so that
// editors with Taplo/Even Better TOML support pick it up automatically.
//
//go:embed schema/forge.schema.json
var SchemaJSON string
