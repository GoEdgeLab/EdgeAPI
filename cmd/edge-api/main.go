package main

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/apps"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/nodes"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	_ "github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/maps"
	"log"
	"os"
)

func main() {
	if !Tea.IsTesting() {
		Tea.Env = "prod"
	}
	app := apps.NewAppCmd()
	app.Version(teaconst.Version)
	app.Product(teaconst.ProductName)
	app.Usage(teaconst.ProcessName + " [start|stop|restart|setup|upgrade|service|daemon]")
	app.On("setup", func() {
		setupCmd := setup.NewSetupFromCmd()
		err := setupCmd.Run()
		result := maps.Map{}
		if err != nil {
			result["isOk"] = false
			result["error"] = err.Error()
		} else {
			result["isOk"] = true
			result["adminNodeId"] = setupCmd.AdminNodeId
			result["adminNodeSecret"] = setupCmd.AdminNodeSecret
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err.Error())
		}

		_, _ = os.Stdout.Write(resultJSON)
	})
	app.On("upgrade", func() {
		fmt.Println("start ...")
		executor, err := setup.NewSQLExecutorFromCmd()
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			return
		}
		err = executor.Run(true)
		if err != nil {
			fmt.Println("ERROR: " + err.Error())
			return
		}
		fmt.Println("finished!")
	})
	app.On("daemon", func() {
		nodes.NewAPINode().Daemon()
	})
	app.On("service", func() {
		err := nodes.NewAPINode().InstallSystemService()
		if err != nil {
			fmt.Println("[ERROR]install failed: " + err.Error())
			return
		}
		fmt.Println("done")
	})
	app.Run(func() {
		nodes.NewAPINode().Start()
	})
}
