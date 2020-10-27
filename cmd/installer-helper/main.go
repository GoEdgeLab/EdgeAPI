package main

import (
	"flag"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"net"
	"os"
)

func main() {
	cmd := ""
	flag.StringVar(&cmd, "cmd", "", "command name: [unzip]")

	// unzip
	zipPath := ""
	targetPath := ""
	flag.StringVar(&zipPath, "zip", "", "zip path")
	flag.StringVar(&targetPath, "target", "", "target dir")

	// parse
	flag.Parse()

	if len(cmd) == 0 {
		stderr("need '-cmd=COMMAND' argument")
	} else if cmd == "test" {
		// 检查是否正在运行
		path := os.TempDir() + "/edge-node.sock"
		conn, err := net.Dial("unix", path)
		if err == nil {
			_ = conn.Close()
			stderr("test node status: edge node is running now, can not install again")
		}
	} else if cmd == "unzip" { // 解压
		if len(zipPath) == 0 {
			stderr("ERROR: need '-zip=PATH' argument")
			return
		}
		if len(targetPath) == 0 {
			stderr("ERROR: need '-target=TARGET' argument")
			return
		}

		unzip := utils.NewUnzip(zipPath, targetPath)
		err := unzip.Run()
		if err != nil {
			stderr("ERROR: " + err.Error())
			return
		}

		stdout("ok")
	} else {
		stderr("ERROR: not recognized command '" + cmd + "'")
	}
}

func stdout(s string) {
	_, _ = os.Stdout.WriteString(s + "\n")
}

func stderr(s string) {
	_, _ = os.Stderr.WriteString(s + "\n")
}
