package setup

import (
	_ "embed"
)

//go:embed sql.json
var sqlData []byte
