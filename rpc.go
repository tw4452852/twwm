package twwm

import (
	"time"
)

var (
	PulseChan = make(chan time.Time)
	Recevers  = []interface{}{
		Ping{},
	}
)

type Ping struct{}

func (Ping) Do(_ struct{}, _ *struct{}) error {
	PulseChan <- time.Now()
	return nil
}
