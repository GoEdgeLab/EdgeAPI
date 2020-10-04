package main

import (
	"github.com/TeaOSLab/EdgeAPI/internal/apps"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/nodes"
	_ "github.com/TeaOSLab/EdgeAPI/internal/tasks"
	_ "github.com/iwind/TeaGo/bootstrap"
)

func main() {
	app := apps.NewAppCmd()
	app.Version(teaconst.Version)
	app.Product(teaconst.ProductName)
	app.Usage(teaconst.ProcessName + " [start|stop|restart]")
	app.Run(func() {
		nodes.NewAPINode().Start()
	})
}
