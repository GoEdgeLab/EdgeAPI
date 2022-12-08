package tasks

import (
	"bytes"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/iwind/TeaGo/maps"
	"time"
)

// HealthCheckClusterTask 单个集群的健康检查任务
type HealthCheckClusterTask struct {
	BaseTask

	clusterId int64
	config    *serverconfigs.HealthCheckConfig
	ticker    *utils.Ticker

	notifiedTime time.Time
}

// NewHealthCheckClusterTask 创建新任务
func NewHealthCheckClusterTask(clusterId int64, config *serverconfigs.HealthCheckConfig) *HealthCheckClusterTask {
	return &HealthCheckClusterTask{
		clusterId: clusterId,
		config:    config,
	}
}

// Reset 重置配置
func (this *HealthCheckClusterTask) Reset(config *serverconfigs.HealthCheckConfig) {
	// 检查是否有变化
	oldJSON, err := json.Marshal(this.config)
	if err != nil {
		this.logErr("HealthCheckClusterTask", err.Error())
		return
	}
	newJSON, err := json.Marshal(config)
	if err != nil {
		this.logErr("HealthCheckClusterTask", err.Error())
		return
	}
	if !bytes.Equal(oldJSON, newJSON) {
		this.config = config
		this.Run()
	}
}

// Run 执行
func (this *HealthCheckClusterTask) Run() {
	this.Stop()

	if this.config == nil {
		return
	}
	if !this.config.IsOn {
		return
	}
	if this.config.Interval == nil {
		return
	}
	var duration = this.config.Interval.Duration()
	if duration <= 0 {
		return
	}
	var ticker = utils.NewTicker(duration)
	goman.New(func() {
		for ticker.Wait() {
			err := this.Loop()
			if err != nil {
				this.logErr("HealthCheckClusterTask", err.Error())
			}
		}
	})
	this.ticker = ticker
}

// Stop 停止
func (this *HealthCheckClusterTask) Stop() {
	if this.ticker == nil {
		return
	}
	this.ticker.Stop()
	this.ticker = nil
}

// Loop 单个循环任务
func (this *HealthCheckClusterTask) Loop() error {
	// 检查是否为主节点
	if !this.IsPrimaryNode() {
		return nil
	}

	// 开始运行
	var executor = NewHealthCheckExecutor(this.clusterId)
	results, err := executor.Run()
	if err != nil {
		return err
	}

	var failedResults = []maps.Map{}
	for _, result := range results {
		if !result.IsOk {
			failedResults = append(failedResults, maps.Map{
				"node": maps.Map{
					"id":   result.Node.Id,
					"name": result.Node.Name,
				},
				"isOk":     false,
				"error":    result.Error,
				"nodeAddr": result.NodeAddr,
			})
		}
	}

	if len(failedResults) > 0 {
		// 10分钟内不重复提醒
		if time.Since(this.notifiedTime) > 10*time.Minute {
			this.notifiedTime = time.Now()

			failedResultsJSON, err := json.Marshal(failedResults)
			if err != nil {
				return err
			}
			var message = "有" + numberutils.FormatInt(len(failedResults)) + "个节点在健康检查中出现问题"
			err = models.NewMessageDAO().CreateClusterMessage(nil, nodeconfigs.NodeRoleNode, this.clusterId, models.MessageTypeHealthCheckFailed, models.MessageLevelError, message, message, failedResultsJSON)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Config 获取当前配置
func (this *HealthCheckClusterTask) Config() *serverconfigs.HealthCheckConfig {
	return this.config
}
