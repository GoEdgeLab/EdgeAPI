package dnsproviders

import "github.com/iwind/TeaGo/maps"

// DNS操作接口
type ProviderInterface interface {
	// 认证
	Auth(params maps.Map) error

	// 读取线路数据
	GetRoutes(domain string) ([][]string, error)
}
