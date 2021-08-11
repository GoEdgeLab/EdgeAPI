package installers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"testing"
)

func TestDNSNodeInstaller_Install(t *testing.T) {
	var installer InstallerInterface = &DNSNodeInstaller{}
	err := installer.Login(&Credentials{
		Host:       "192.168.2.30",
		Port:       22,
		Username:   "root",
		Password:   "123456",
		PrivateKey: "",
		Method:     "user",
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
		Endpoints: []string{"http://192.168.2.40:8003"},
		NodeId:    "b3f0690c793db5daaa666e89bd7b2301",
		Secret:    "H6nbSzjN3tLYi0ecdtUeDpQdZZPjKL7S",
	}, &models.NodeInstallStatus{})
	if err != nil {
		t.Fatal(err)
	}
}
