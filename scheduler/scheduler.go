package scheduler

import (
	"sync"
	"time"
)

type Worker func()

type Scheduler struct {
	interval    time.Duration
	waitGroup   sync.WaitGroup
	worker      Worker
	stopChannel chan int
}

func Create(worker Worker, interval time.Duration) *Scheduler {
	return &Scheduler{
		worker:      worker,
		stopChannel: make(chan int, 1),
		interval:    interval,
	}
}

func (s *Scheduler) Start() {
	s.waitGroup.Add(1)
	go s.schedule()
}

func (s *Scheduler) Stop() {
	s.stopChannel <- 0
	s.waitGroup.Wait()
}

func (s *Scheduler) schedule() {
	timer := time.NewTimer(time.Nanosecond)
	for {
		select {
		case <-timer.C:
			s.runWorker(timer)

		case <-s.stopChannel:
			s.waitGroup.Done()
			return
		}
	}
}

func (s *Scheduler) runWorker(timer *time.Timer) {
	startTime := time.Now()
	s.worker()
	passedTime := time.Now().Sub(startTime)
	waitTime := s.interval - passedTime
	if waitTime < 0 {
		waitTime = time.Nanosecond
	}
	timer.Reset(waitTime)
}
