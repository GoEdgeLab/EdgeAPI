package tasks

import (
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/logs"
	"time"
)

func init() {
	looper := NewEventLooper()
	go looper.Start()
}

type EventLooper struct {
}

func NewEventLooper() *EventLooper {
	return &EventLooper{}
}

func (this *EventLooper) Start() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		err := this.loop()
		if err != nil {
			logs.Println("[EVENT_LOOPER]" + err.Error())
		}
	}
}

func (this *EventLooper) loop() error {
	lockerKey := "eventLooper"
	isOk, err := models.SharedSysLockerDAO.Lock(lockerKey, 3600)
	if err != nil {
		return err
	}
	defer func() {
		err = models.SharedSysLockerDAO.Unlock(lockerKey)
		if err != nil {
			logs.Println("[EVENT_LOOPER]" + err.Error())
		}
	}()
	if !isOk {
		return nil
	}

	events, err := models.SharedSysEventDAO.FindEvents(100)
	if err != nil {
		return err
	}
	for _, eventOne := range events {
		event, err := eventOne.DecodeEvent()
		if err != nil {
			logs.Println("[EVENT_LOOPER]" + err.Error())
			continue
		}
		err = event.Run()
		if err != nil {
			logs.Println("[EVENT_LOOPER]" + err.Error())
			continue
		}
		err = models.SharedSysEventDAO.DeleteEvent(int64(eventOne.Id))
		if err != nil {
			return err
		}
	}

	return nil
}
