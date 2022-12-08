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
			NewEventLooper(2 * time.Second).Start()
		})
	})
}

// EventLooper 事件相关处理程序
type EventLooper struct {
	BaseTask

	ticker *time.Ticker
}

func NewEventLooper(duration time.Duration) *EventLooper {
	return &EventLooper{
		ticker: time.NewTicker(duration),
	}
}

func (this *EventLooper) Start() {
	for range this.ticker.C {
		err := this.Loop()
		if err != nil {
			this.logErr("EventLooper", err.Error())
		}
	}
}

func (this *EventLooper) Loop() error {
	if !this.IsPrimaryNode() {
		return nil
	}

	events, err := models.SharedSysEventDAO.FindEvents(nil, 100)
	if err != nil {
		return err
	}
	for _, eventOne := range events {
		event, err := eventOne.DecodeEvent()
		if err != nil {
			this.logErr("EventLooper", err.Error())
			continue
		}
		err = event.Run()
		if err != nil {
			this.logErr("EventLooper", err.Error())
			continue
		}
		err = models.SharedSysEventDAO.DeleteEvent(nil, int64(eventOne.Id))
		if err != nil {
			return err
		}
	}

	return nil
}
