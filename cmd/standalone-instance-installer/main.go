// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package main

import (
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/instances"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/lists"
	"os"
	"strings"
)

func main() {
	dbPasswordData, err := os.ReadFile("/usr/local/mysql/generated-password.txt")
	if err != nil {
		fmt.Println("[ERROR]read mysql password failed: " + err.Error())
		return
	}
	var dbPassword = strings.TrimSpace(string(dbPasswordData))

	var isTesting = lists.ContainsString(os.Args, "-test") || lists.ContainsString(os.Args, "--test")
	if isTesting {
		fmt.Println("testing mode ...")
	}

	var instance = instances.NewInstance(instances.Options{
		IsTesting: isTesting,
		Verbose:   lists.ContainsString(os.Args, "-v"),
		Cacheable: true,
		WorkDir:   "",
		SrcDir:    "/usr/local/goedge/src",
		DB: struct {
			Host     string
			Port     int
			Username string
			Password string
			Name     string
		}{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "root",
			Password: dbPassword,
			Name:     "edges",
		},
		AdminNode: struct {
			Port int
		}{
			Port: 7788,
		},
		APINode: struct {
			HTTPPort     int
			RestHTTPPort int
		}{
			HTTPPort:     8001,
			RestHTTPPort: 8002,
		},
		Node: struct{ HTTPPort int }{
			HTTPPort: 8080,
		},
		UserNode: struct {
			HTTPPort int
		}{
			HTTPPort: 7799,
		},
	})
	err = instance.SetupAll()
	if err != nil {
		fmt.Println("[ERROR]setup failed: " + err.Error())
		return
	}

	fmt.Println("ok")
}
