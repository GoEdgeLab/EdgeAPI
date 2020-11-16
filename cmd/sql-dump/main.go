package main

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	db, err := dbs.Default()
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}
	results, err := setup.NewSQLDump().Dump(db)
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}
	dir, _ := os.Getwd()
	var sqlFile string
	for i := 0; i < 5; i++ {
		lookupFile := dir + "/internal/setup/sql.go"
		_, err = os.Stat(lookupFile)
		if err != nil {
			dir = filepath.Dir(dir)
			continue
		}
		sqlFile = lookupFile
	}

	if len(sqlFile) == 0 {
		fmt.Println("[ERROR]can not find sql.go")
		return
	}
	content := []byte(`package setup

import (
	"encoding/json"
	"github.com/iwind/TeaGo/logs"
)

// 最新的SQL语句
// 由sql-dump/main.go自动生成

func init() {
	err := json.Unmarshal([]byte(` + strconv.Quote(string(resultsJSON)) + `), LatestSQLResult)
	if err != nil {
		logs.Println("[ERROR]load sql failed: " + err.Error())
	}
}
`)
	dst, err := format.Source(content)
	if err != nil {
		fmt.Println("[ERROR]format code failed: " + err.Error())
		return
	}

	err = ioutil.WriteFile(sqlFile, dst, 0666)
	if err != nil {
		fmt.Println("[ERROR]write file failed: " + err.Error())
		return
	}
	fmt.Println("ok")
}
