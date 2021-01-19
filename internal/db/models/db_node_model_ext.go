package models

import (
	"fmt"
	"github.com/iwind/TeaGo/dbs"
)

// 获取数据库配置
func (this *DBNode) DBConfig() *dbs.DBConfig {
	dsn := this.Username + ":" + this.Password + "@tcp(" + this.Host + ":" + fmt.Sprintf("%d", this.Port) + ")/" + this.Database + "?charset=utf8mb4&timeout=10s"

	return &dbs.DBConfig{
		Driver: "mysql",
		Dsn:    dsn,
		Prefix: "edge",
	}
}
