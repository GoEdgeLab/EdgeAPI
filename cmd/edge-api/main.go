package main

import (
	"github.com/TeaOSLab/EdgeAPI/internal/apis"
	_ "github.com/iwind/TeaGo/bootstrap"
)

func main() {
	apis.NewAPINode().Start()
}
