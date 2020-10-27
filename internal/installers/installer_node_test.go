package installers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"testing"
)

func TestNodeInstaller_Install(t *testing.T) {
	var installer InstallerInterface = &NodeInstaller{}
	err := installer.Login(&Credentials{
		Host:       "192.168.2.30",
		Port:       22,
		Username:   "root",
		Password:   "123456",
		PrivateKey: "",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 关闭连接
	defer func() {
		err := installer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 安装
	err = installer.Install("/opt/edge", &NodeParams{
		Endpoints: []string{"192.168.2.40:8003"},
		NodeId:    "313fdb1b90d0a63c736f307b4d1ca358",
		Secret:    "Pl3u5kYqBDZddp7raw6QfHiuGPRCWF54",
	}, &models.NodeInstallStatus{})
	if err != nil {
		t.Fatal(err)
	}
}
