package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/numberutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewLogTask(24*time.Hour, 1*time.Minute).Start()
		})
	})
}

type LogTask struct {
	BaseTask

	cleanTicker   *time.Ticker
	monitorTicker *time.Ticker
}

func NewLogTask(cleanDuration time.Duration, monitorDuration time.Duration) *LogTask {
	return &LogTask{
		cleanTicker:   time.NewTicker(cleanDuration),
		monitorTicker: time.NewTicker(monitorDuration),
	}
}

func (this *LogTask) Start() {
	goman.New(func() {
		this.RunClean()
	})
	goman.New(func() {
		this.RunMonitor()
	})
}

func (this *LogTask) RunClean() {
	for range this.cleanTicker.C {
		err := this.LoopClean()
		if err != nil {
			this.logErr("LogTask", err.Error())
		}
	}
}

func (this *LogTask) LoopClean() error {
	var configKey = "adminLogConfig"
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(nil, configKey)
	if err != nil {
		return err
	}
	if len(valueJSON) == 0 {
		return nil
	}

	var config = &systemconfigs.LogConfig{}
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

func (this *LogTask) RunMonitor() {
	for range this.monitorTicker.C {
		err := this.LoopMonitor()
		if err != nil {
			this.logErr("LogTask", err.Error())
		}
	}
}

func (this *LogTask) LoopMonitor() error {
	// 检查是否为主节点
	if !this.IsPrimaryNode() {
		return nil
	}

	var configKey = "adminLogConfig"
	valueJSON, err := models.SharedSysSettingDAO.ReadSetting(nil, configKey)
	if err != nil {
		return err
	}
	if len(valueJSON) == 0 {
		return nil
	}

	var config = &systemconfigs.LogConfig{}
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
				err := models.SharedMessageDAO.CreateMessage(nil, 0, 0, models.MessageTypeLogCapacityOverflow, models.MessageLevelError, "日志用量已经超出最大限制", "日志用量已经超出最大限制，当前的用量为"+this.formatBytes(sumBytes)+"，而设置的最大容量为"+this.formatBytes(capacityBytes)+"。", nil)
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
