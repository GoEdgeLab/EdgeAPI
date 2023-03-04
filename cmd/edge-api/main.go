package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/apps"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/nodes"
	"github.com/TeaOSLab/EdgeAPI/internal/setup"
	_ "github.com/TeaOSLab/EdgeAPI/internal/tasks"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/gosock/pkg/gosock"
	"log"
	"os"
)

func main() {
	if !Tea.IsTesting() {
		Tea.Env = "prod"
	}
	var app = apps.NewAppCmd()
	app.Version(teaconst.Version)
	app.Product(teaconst.ProductName)
	app.Usage(teaconst.ProcessName + " [-h|-v|start|stop|restart|setup|upgrade|service|daemon|issues]")

	// 短版本号
	app.On("-V", func() {
		_, _ = os.Stdout.WriteString(teaconst.Version)
	})
	app.On("setup", func() {
		var setupCmd = setup.NewSetupFromCmd()
		err := setupCmd.Run()
		var result = maps.Map{}
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
	app.On("reset", func() {
		err := configs.ResetAPIConfig()
		if err != nil {
			fmt.Println("[ERROR]reset failed: " + err.Error())
			return
		}
		fmt.Println("done")
	})
	app.On("goman", func() {
		var sock = gosock.NewTmpSock(teaconst.ProcessName)
		reply, err := sock.Send(&gosock.Command{Code: "goman"})
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
		} else {
			instancesJSON, err := json.MarshalIndent(reply.Params, "", "  ")
			if err != nil {
				fmt.Println("[ERROR]" + err.Error())
			} else {
				fmt.Println(string(instancesJSON))
			}
		}
	})
	app.On("debug", func() {
		var sock = gosock.NewTmpSock(teaconst.ProcessName)
		reply, err := sock.Send(&gosock.Command{Code: "debug"})
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
		} else {
			var isDebug = maps.NewMap(reply.Params).GetBool("debug")
			if isDebug {
				fmt.Println("debug on")
			} else {
				fmt.Println("debug off")
			}
		}
	})
	app.On("db.stmt.prepare", func() {
		var sock = gosock.NewTmpSock(teaconst.ProcessName)
		reply, err := sock.Send(&gosock.Command{Code: "db.stmt.prepare"})
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
		} else {
			var isOn = maps.NewMap(reply.Params).GetBool("isOn")
			if isOn {
				fmt.Println("show statements: on")
			} else {
				fmt.Println("show statements: off")
			}
		}
	})
	app.On("db.stmt.count", func() {
		var sock = gosock.NewTmpSock(teaconst.ProcessName)
		reply, err := sock.Send(&gosock.Command{Code: "db.stmt.count"})
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
		} else {
			var count = maps.NewMap(reply.Params).GetInt("count")
			fmt.Println("prepared statements count: " + types.String(count))
		}
	})
	app.On("issues", func() {
		var flagSet = flag.NewFlagSet("issues", flag.ExitOnError)
		var formatJSON = false
		flagSet.BoolVar(&formatJSON, "json", false, "")
		_ = flagSet.Parse(os.Args[2:])

		data, err := os.ReadFile(Tea.LogFile("issues.log"))
		if err != nil {
			if formatJSON {
				fmt.Print("[]")
			} else {
				fmt.Println("no issues yet")
			}
		} else {
			var issueMaps = []maps.Map{}
			err = json.Unmarshal(data, &issueMaps)
			if err != nil {
				if formatJSON {
					fmt.Print("[]")
				} else {
					fmt.Println("no issues yet")
				}
			} else {
				if formatJSON {
					fmt.Print(string(data))
				} else {
					if len(issueMaps) == 0 {
						fmt.Println("no issues yet")
					} else {
						for i, issue := range issueMaps {
							fmt.Println("issue " + types.String(i+1) + ": " + issue.GetString("message"))
						}
					}
				}
			}
		}
	})
	app.On("instance", func() {
		var sock = gosock.NewTmpSock(teaconst.ProcessName)
		reply, err := sock.Send(&gosock.Command{Code: "instance"})
		if err != nil {
			fmt.Println("[ERROR]" + err.Error())
		} else {
			replyJSON, err := json.MarshalIndent(reply.Params, "", "  ")
			if err != nil {
				fmt.Println("[ERROR]marshal result failed: " + err.Error())
			} else {
				fmt.Println(string(replyJSON))
			}
		}
	})

	app.Run(func() {
		nodes.NewAPINode().Start()
	})
}
