package tasks

import (
	"bytes"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"time"
)

// 单个集群的健康检查任务
type HealthCheckClusterTask struct {
	clusterId int64
	config    *serverconfigs.HealthCheckConfig
	ticker    *utils.Ticker
}

// 创建新任务
func NewHealthCheckClusterTask(clusterId int64, config *serverconfigs.HealthCheckConfig) *HealthCheckClusterTask {
	return &HealthCheckClusterTask{
		clusterId: clusterId,
		config:    config,
	}
}

// 重置配置
func (this *HealthCheckClusterTask) Reset(config *serverconfigs.HealthCheckConfig) {
	// 检查是否有变化
	oldJSON, err := json.Marshal(this.config)
	if err != nil {
		logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
		return
	}
	newJSON, err := json.Marshal(config)
	if err != nil {
		logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
		return
	}
	if bytes.Compare(oldJSON, newJSON) != 0 {
		this.config = config
		this.Run()
	}
}

// 执行
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
	duration := this.config.Interval.Duration()
	if duration <= 0 {
		return
	}
	ticker := utils.NewTicker(duration)
	go func() {
		for ticker.Wait() {
			err := this.loop(int64(duration.Seconds()))
			if err != nil {
				logs.Println("[TASK][HEALTH_CHECK]" + err.Error())
			}
		}
	}()
	this.ticker = ticker
}

// 停止
func (this *HealthCheckClusterTask) Stop() {
	if this.ticker == nil {
		return
	}
	this.ticker.Stop()
	this.ticker = nil
}

// 单个循环任务
func (this *HealthCheckClusterTask) loop(seconds int64) error {
	// 检查上次运行时间，防止重复运行
	settingKey := systemconfigs.SettingCodeClusterHealthCheck + "Loop" + numberutils.FormatInt64(this.clusterId)
	timestamp := time.Now().Unix()
	c, err := models.SharedSysSettingDAO.CompareInt64Setting(nil, settingKey, timestamp-seconds)
	if err != nil {
		return err
	}
	if c > 0 {
		return nil
	}

	// 记录时间
	err = models.SharedSysSettingDAO.UpdateSetting(nil, settingKey, []byte(numberutils.FormatInt64(timestamp)))
	if err != nil {
		return err
	}

	// 开始运行
	executor := NewHealthCheckExecutor(this.clusterId)
	results, err := executor.Run()
	if err != nil {
		return err
	}

	failedResults := []maps.Map{}
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
		failedResultsJSON, err := json.Marshal(failedResults)
		if err != nil {
			return err
		}
		message := "有" + numberutils.FormatInt(len(failedResults)) + "个节点在健康检查中出现问题"
		err = models.NewMessageDAO().CreateClusterMessage(nil, this.clusterId, models.MessageTypeHealthCheckFailed, models.MessageLevelError, message, message, failedResultsJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// Config 获取当前配置
func (this *HealthCheckClusterTask) Config() *serverconfigs.HealthCheckConfig {
	return this.config
}
