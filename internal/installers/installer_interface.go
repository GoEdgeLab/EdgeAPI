package installers

type InstallerInterface interface {
	// 登录SSH服务
	Login(credentials *Credentials) error

	// 安装
	Install(dir string, params interface{}) error

	// 关闭连接的SSH服务
	Close() error
}
