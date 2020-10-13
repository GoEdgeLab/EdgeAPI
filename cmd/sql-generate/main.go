package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/logs"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", "", "SQL dir")
	flag.Parse()

	if len(dir) == 0 {
		fmt.Println("[ERROR]'dir' should not be empty")
		return
	}

	sourceDir := filepath.Dir(Tea.Root)

	// full
	fullSQLFile := dir + "/full.sql"
	_, err := os.Stat(fullSQLFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[ERROR]'full.sql' not found")
			return
		}
		fmt.Println("[ERROR]checking 'full.sql' failed: " + err.Error())
		return
	}

	matches, err := filepath.Glob(dir + "/*.sql")
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}

	versionsCode := "// generated\npackage sqls\n\nvar SQLVersions = []map[string]string{"

	for _, match := range matches {
		baseName := filepath.Base(match)
		logs.Println("reading " + baseName + " ...")

		data, err := ioutil.ReadFile(match)
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
			return
		}

		version := baseName[:strings.LastIndex(baseName, ".")]
		versionsCode += "\n" + `{ "version": "` + version + `", "sql": SQL_` + version + ` },`

		code := "// generated\npackage sqls \n\n"
		lines := bytes.Split(data, []byte{'\n'})
		for index, line := range lines {
			if index == 0 {
				code += "var SQL_" + version + " = "
			}
			code += `"` + string(line) + `\n"`
			if index != len(lines)-1 {
				code += "+"
			}
			code += "\n"
		}
		code += "\n"

		codeBytes, err := format.Source([]byte(code))
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
			return
		}
		fmt.Println("writing sql_" + version + ".go ...")
		err = ioutil.WriteFile(sourceDir+"/internal/setup/sqls/sql_"+version+".go", codeBytes, 0666)
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
			return
		}
		fmt.Println("ok")
	}

	versionsCode += "\n}"
	versionsCodeBytes, err := format.Source([]byte(versionsCode))
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}
	fmt.Println("writing sqls.go ...")
	err = ioutil.WriteFile(sourceDir+"/internal/setup/sqls/sqls.go", versionsCodeBytes, 0666)
	if err != nil {
		fmt.Println("[ERROR]" + err.Error())
		return
	}
	fmt.Println("ok")
}
