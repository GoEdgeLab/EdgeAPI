package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/iwind/TeaGo/dbs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewMessageTask(24 * time.Hour).Start()
		})
	})
}

// MessageTask 消息相关任务
type MessageTask struct {
	BaseTask

	ticker *time.Ticker
}

// NewMessageTask 获取新对象
func NewMessageTask(duration time.Duration) *MessageTask {
	return &MessageTask{
		ticker: time.NewTicker(duration),
	}
}

// Start 开始运行
func (this *MessageTask) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("MessageTask", err.Error())
		}
	}
}

// Loop 单次运行
func (this *MessageTask) Loop() error {
	dayTime := time.Now().AddDate(0, 0, -30) // TODO 这个30天应该可以在界面上设置
	return models.NewMessageDAO().DeleteMessagesBeforeDay(nil, dayTime)
}
