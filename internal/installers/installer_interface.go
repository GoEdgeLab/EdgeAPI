package installers

import "github.com/TeaOSLab/EdgeAPI/internal/db/models"

type InstallerInterface interface {
	// 登录SSH服务
	Login(credentials *Credentials) error

	// 安装
	Install(dir string, params interface{}, installStatus *models.NodeInstallStatus) error

	// 关闭连接的SSH服务
	Close() error
}
