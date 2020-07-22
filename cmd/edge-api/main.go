package main

import (
	"github.com/TeaOSLab/EdgeAPI/internal/apis"
	"github.com/TeaOSLab/EdgeAPI/internal/apps"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	_ "github.com/iwind/TeaGo/bootstrap"
)

func main() {
	app := apps.NewAppCmd()
	app.Version(teaconst.Version)
	app.Product(teaconst.ProductName)
	app.Usage(teaconst.ProcessName + " [start|stop|restart]")
	app.Run(func() {
		apis.NewAPINode().Start()
	})
}
