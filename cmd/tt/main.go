package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/cmd"
	_ "github.com/iwind/TeaGo/dbs/commands"
	"github.com/iwind/TeaGo/lists"
	"os"
	"path/filepath"
	"time"
)

// TeaTool工具
func main() {
	r := bufio.NewReader(os.Stdin)
	lastCommand := ""

	for {
		time.Sleep(400 * time.Millisecond)
		fmt.Print("> ")

		line, _, err := r.ReadLine()
		if err != nil {
			continue
		}

		command := string(bytes.TrimSpace(line))

		// 命令帮助
		if len(command) == 0 || command == "help" || command == "h" || command == "?" || command == "/?" {
			lastCommand = command
			fmt.Println("TeaTool commands:")
			commands := cmd.AllCommands()

			// 对命令代码进行排序
			codes := []string{}
			for code := range commands {
				codes = append(codes, code)
			}

			lists.Sort(codes, func(i int, j int) bool {
				code1 := codes[i]
				code2 := codes[j]
				return code1 < code2
			})

			//输出
			for _, code := range codes {
				ptr := commands[code]
				fmt.Println("  ", code+"\n\t\t"+ptr.Name())
			}
			continue
		}

		if command == "retry" || command == "!!" /** csh like **/ || command == "!-1" /** csh like **/ {
			command = lastCommand
			fmt.Println("retry '" + command + "'")
		}
		lastCommand = command

		found := cmd.Try(cmd.ParseArgs(command))
		if !found {
			fmt.Println("command '" + command + "' not found")
		}
	}
}

// 重置Root
func init() {
	webIsSet := false
	if !Tea.IsTesting() {
		exePath, err := os.Executable()
		if err != nil {
			exePath = os.Args[0]
		}
		link, err := filepath.EvalSymlinks(exePath)
		if err == nil {
			exePath = link
		}
		fullPath, err := filepath.Abs(exePath)
		if err == nil {
			Tea.UpdateRoot(filepath.Dir(filepath.Dir(fullPath)))
		}
	} else {
		pwd, ok := os.LookupEnv("PWD")
		if ok {
			webIsSet = true
			Tea.SetPublicDir(pwd + Tea.DS + "web" + Tea.DS + "public")
			Tea.SetViewsDir(pwd + Tea.DS + "web" + Tea.DS + "views")
			Tea.SetTmpDir(pwd + Tea.DS + "web" + Tea.DS + "tmp")

			Tea.Root = pwd + Tea.DS + "build"
		}
	}

	if !webIsSet {
		Tea.SetPublicDir(Tea.Root + Tea.DS + "web" + Tea.DS + "public")
		Tea.SetViewsDir(Tea.Root + Tea.DS + "web" + Tea.DS + "views")
		Tea.SetTmpDir(Tea.Root + Tea.DS + "web" + Tea.DS + "tmp")
	}
	Tea.SetConfigDir(Tea.Root + Tea.DS + "configs")

	_ = os.Setenv("GOPATH", filepath.Dir(Tea.Root))
}
