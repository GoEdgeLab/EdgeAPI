package utils

import "time"

type Ticker struct {
	raw    *time.Ticker
	isDone bool
	done   chan bool
}

func NewTicker(duration time.Duration) *Ticker {
	return &Ticker{
		raw:  time.NewTicker(duration),
		done: make(chan bool),
	}
}

func (this *Ticker) Wait() bool {
	select {
	case <-this.raw.C:
		return true
	case <-this.done:
		this.isDone = true
		return false
	}
}

func (this *Ticker) Stop() {
	if this.isDone {
		return
	}
	this.done <- true
	this.raw.Stop()
}
