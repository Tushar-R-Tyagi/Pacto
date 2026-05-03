package timer

import "time"

type BoutTimer struct {
	SecondsLeft int
	Running     bool
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
	t.SecondsLeft = totalSeconds
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				t.SecondsLeft--
				select {
				case t.Tick <- t.SecondsLeft:
				default:
				}
				if t.SecondsLeft <= 0 {
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
