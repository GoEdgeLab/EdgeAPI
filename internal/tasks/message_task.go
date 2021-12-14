package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			NewMessageTask().Run()
		})
	})
}

// MessageTask 消息相关任务
type MessageTask struct {
}

// NewMessageTask 获取新对象
func NewMessageTask() *MessageTask {
	return &MessageTask{}
}

// Run 运行
func (this *MessageTask) Run() {
	ticker := utils.NewTicker(24 * time.Hour)
	for ticker.Wait() {
		err := this.loop()
		if err != nil {
			logs.Println("[TASK][MESSAGE]" + err.Error())
		}
	}
}

// 单次运行
func (this *MessageTask) loop() error {
	dayTime := time.Now().AddDate(0, 0, -30) // TODO 这个30天应该可以在界面上设置
	return models.NewMessageDAO().DeleteMessagesBeforeDay(nil, dayTime)
}
