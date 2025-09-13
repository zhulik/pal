package inspect

import (
	"embed"
	_ "embed"
)

//go:embed static
var StaticFS embed.FS
