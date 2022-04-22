// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package accesslogs

import "github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"

// StorageInterface 日志存储接口
type StorageInterface interface {
	// Version 获取版本
	Version() int

	// SetVersion 设置版本
	SetVersion(version int)

	// SetFirewallOnly 设置是否只处理防火墙相关的访问日志
	SetFirewallOnly(firewallOnly bool)

	IsOk() bool

	SetOk(ok bool)

	// Config 获取配置
	Config() interface{}

	// Start 开启
	Start() error

	// Write 写入日志
	Write(accessLogs []*pb.HTTPAccessLog) error

	// Close 关闭
	Close() error
}
