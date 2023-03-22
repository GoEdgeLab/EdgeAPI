package setup

import (
	_ "embed"
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

//go:embed sql.json
var sqlData []byte

func init() {
	err := json.Unmarshal(sqlData, LatestSQLResult)
	if err != nil {
		logs.Println("[ERROR]load sql failed: " + err.Error())
	}
}
