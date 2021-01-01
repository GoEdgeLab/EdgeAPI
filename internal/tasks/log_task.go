package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReady(func() {
		go NewLogTask().Run()
	})
}

type LogTask struct {
}

func NewLogTask() *LogTask {
	return &LogTask{}
}

func (this *LogTask) Run() {
	go this.runClean()
	go this.runMonitor()
}

func (this *LogTask) runClean() {
	ticker := utils.NewTicker(24 * time.Hour)
	for ticker.Wait() {
		err := this.loopClean(86400)
		if err != nil {
			logs.Println("[TASK][LOG]" + err.Error())
		}
	}
}

func (this *LogTask) loopClean(seconds int64) error {
	// 检查上次运行时间，防止重复运行
	settingKey := "logTaskCleanLoop"
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

	configKey := "adminLogConfig"
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(nil, configKey)
	if err != nil {
		return err
	}
	if len(valueJSON) == 0 {
		return nil
	}

	config := &systemconfigs.LogConfig{}
	err = json.Unmarshal(valueJSON, config)
	if err != nil {
		return err
	}
	if config.Days > 0 {
		err = models.SharedLogDAO.DeleteLogsPermanentlyBeforeDays(nil, config.Days)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *LogTask) runMonitor() {
	ticker := utils.NewTicker(1 * time.Minute)
	for ticker.Wait() {
		err := this.loopMonitor(60)
		if err != nil {
			logs.Println("[TASK][LOG]" + err.Error())
		}
	}
}

func (this *LogTask) loopMonitor(seconds int64) error {
	// 检查上次运行时间，防止重复运行
	settingKey := "logTaskMonitorLoop"
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

	configKey := "adminLogConfig"
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(nil, configKey)
	if err != nil {
		return err
	}
	if len(valueJSON) == 0 {
		return nil
	}

	config := &systemconfigs.LogConfig{}
	err = json.Unmarshal(valueJSON, config)
	if err != nil {
		return err
	}

	if config.Capacity != nil {
		capacityBytes := config.Capacity.Bytes()
		if capacityBytes > 0 {
			sumBytes, err := models.SharedLogDAO.SumLogsSize()
			if err != nil {
				return err
			}
			if sumBytes > capacityBytes {
				err := models.SharedMessageDAO.CreateMessage(nil, 0, 0, models.MessageTypeLogCapacityOverflow, models.MessageLevelError, "日志用量已经超出最大限制，当前的用量为"+this.formatBytes(sumBytes)+"，而设置的最大容量为"+this.formatBytes(capacityBytes)+"。", nil)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (this *LogTask) formatBytes(bytes int64) string {
	sizeHuman := ""
	if bytes < 1024 {
		sizeHuman = numberutils.FormatInt64(bytes) + "字节"
	} else if bytes < 1024*1024 {
		sizeHuman = fmt.Sprintf("%.2fK", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		sizeHuman = fmt.Sprintf("%.2fM", float64(bytes)/1024/1024)
	} else {
		sizeHuman = fmt.Sprintf("%.2fG", float64(bytes)/1024/1024/1024)
	}
	return sizeHuman
}
