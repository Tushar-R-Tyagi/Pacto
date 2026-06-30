package timer

import "time"

type BoutTimer struct {
	SecondsLeft int
	Running     bool
	endTime     time.Time      // ← the anchor: when the bout should finish
	Done        chan struct{}
	Tick        chan int
	Cancel      chan struct{}
}

func New() *BoutTimer {
	return &BoutTimer{
		Done:   make(chan struct{}),
		Tick:   make(chan int),
		Cancel: make(chan struct{}),
	}
}

func (t *BoutTimer) Start(totalSeconds int) {
	t.Running = true
	t.endTime = time.Now().Add(time.Duration(totalSeconds) * time.Second) // anchor to real clock
	t.SecondsLeft = totalSeconds

	ticker := time.NewTicker(250 * time.Millisecond) // poll often; accuracy comes from endTime, not tick count
	go func() {
		for {
			select {
			case <-ticker.C:
				// derive remaining time from the system clock, not by counting
				remaining := int(time.Until(t.endTime).Seconds() + 0.5) // +0.5 rounds to nearest second
				if remaining < 0 {
					remaining = 0
				}
				t.SecondsLeft = remaining

				select {
				case t.Tick <- t.SecondsLeft:
				default:
				}

				if remaining <= 0 {
					t.Running = false
					ticker.Stop()
					t.Done <- struct{}{}
					return
				}
			case <-t.Cancel:
				t.Running = false
				ticker.Stop()
				return
			}
		}
	}()
}

func (t *BoutTimer) Stop() {
	if t.Running {
		t.Cancel <- struct{}{}
	}
}
