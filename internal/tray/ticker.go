package tray

import "time"

// ticker wraps time.Ticker so the poll loop can be stopped cleanly.
type ticker struct {
	t  *time.Ticker
	ch <-chan time.Time
}

func newTicker(d time.Duration) *ticker {
	t := time.NewTicker(d)
	return &ticker{t: t, ch: t.C}
}

func (t *ticker) stop() {
	t.t.Stop()
}
