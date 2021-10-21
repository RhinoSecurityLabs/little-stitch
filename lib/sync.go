package lib

import (
	"sync"
	"time"
)

type Syncer struct {
	Delay time.Duration
	Functions []func()
}

func (s *Syncer) Start() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	go func() {
		for _, f := range s.Functions {
			wg.Add(1)
			go func(f func()) {
				f()
				wg.Done()
			}(f)
		}
		time.Sleep(s.Delay)
	}()
	return wg
}
