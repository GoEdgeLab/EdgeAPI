package models

// NSCluster 域名服务器集群
type NSCluster struct {
	Id         uint32 `field:"id"`         // ID
	IsOn       uint8  `field:"isOn"`       // 是否启用
	Name       string `field:"name"`       // 集群名
	InstallDir string `field:"installDir"` // 安装目录
	State      uint8  `field:"state"`      // 状态
	AccessLog  string `field:"accessLog"`  // 访问日志配置
	GrantId    uint32 `field:"grantId"`    // 授权ID
	Recursion  string `field:"recursion"`  // 递归DNS设置
}

type NSClusterOperator struct {
	Id         interface{} // ID
	IsOn       interface{} // 是否启用
	Name       interface{} // 集群名
	InstallDir interface{} // 安装目录
	State      interface{} // 状态
	AccessLog  interface{} // 访问日志配置
	GrantId    interface{} // 授权ID
	Recursion  interface{} // 递归DNS设置
}

func NewNSClusterOperator() *NSClusterOperator {
	return &NSClusterOperator{}
}
