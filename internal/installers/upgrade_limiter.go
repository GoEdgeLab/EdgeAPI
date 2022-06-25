// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package installers

import (
	"github.com/TeaOSLab/EdgeAPI/internal/utils/sizes"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/iwind/TeaGo/types"
	"sync"
	"time"
)

const (
	UpgradeLimiterDuration          = 10          // node key expire time, by seconds
	UpgradeLimiterConcurrent        = 10          // 10 nodes
	UpgradeLimiterMaxBytesPerSecond = 5 * sizes.M // max bytes per second
)

var SharedUpgradeLimiter = NewUpgradeLimiter()

// UpgradeLimiter 升级流量管理器
type UpgradeLimiter struct {
	nodeMap map[string]int64 // key => timestamp

	rateTimestamp int64
	rateBytes     int64

	locker sync.Mutex
}

func NewUpgradeLimiter() *UpgradeLimiter {
	return &UpgradeLimiter{
		nodeMap: map[string]int64{},
	}
}

// UpdateNodeBytes 添加正在下载的节点流量
func (this *UpgradeLimiter) UpdateNodeBytes(nodeType nodeconfigs.NodeRole, nodeId int64, bytes int64) {
	this.locker.Lock()
	defer this.locker.Unlock()

	// 先清理
	var nowTime = time.Now().Unix()
	this.gc(nowTime)

	// 添加
	var key = nodeType + "_" + types.String(nodeId)
	this.nodeMap[key] = nowTime

	// 流量
	if this.rateTimestamp == nowTime {
		this.rateBytes += bytes
	} else {
		this.rateTimestamp = nowTime
		this.rateBytes = bytes
	}
}

// CanUpgrade 检查是否有新的升级
func (this *UpgradeLimiter) CanUpgrade() bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	var nowTime = time.Now().Unix()
	this.gc(nowTime)

	// 限制并发节点数
	if len(this.nodeMap) >= UpgradeLimiterConcurrent {
		return false
	}

	if this.rateTimestamp != nowTime {
		return true
	}

	// 限制下载速度
	if this.rateBytes >= UpgradeLimiterMaxBytesPerSecond {
		return false
	}

	return true
}

func (this *UpgradeLimiter) gc(nowTime int64) {
	for nodeKey, timestamp := range this.nodeMap {
		if timestamp < nowTime-UpgradeLimiterDuration {
			delete(this.nodeMap, nodeKey)
		}
	}
}
