package core

import (
	"sync"
	"sync/atomic"
	"time"
)

// PollInterval manages a thread-safe, dynamic polling interval.
type PollInterval struct {
	nanos atomic.Int64
}

// NewPollInterval creates a PollInterval with the given initial duration.
func NewPollInterval(initial time.Duration) *PollInterval {
	p := &PollInterval{}
	p.Set(initial)
	return p
}

// Duration returns the current interval as a time.Duration.
func (p *PollInterval) Duration() time.Duration {
	return time.Duration(p.nanos.Load())
}

// Set dynamically updates the interval.
func (p *PollInterval) Set(d time.Duration) {
	if d <= 0 {
		d = time.Second
	}
	p.nanos.Store(d.Nanoseconds())
}

// PollLoop manages a self-rescheduling timer loop with a dynamic delay function.
type PollLoop struct {
	run   func()
	delay func() time.Duration

	mu      sync.Mutex
	stopped bool
	timer   *time.Timer
	lastRun time.Time
}

// NewPollLoop creates a PollLoop that executes run periodically.
func NewPollLoop(run func(), delay func() time.Duration) *PollLoop {
	return &PollLoop{run: run, delay: delay, stopped: true}
}

// Start starts the polling loop if not already running.
func (p *PollLoop) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.stopped {
		return
	}
	p.stopped = false
	wait := p.delay()
	if wait <= 0 {
		wait = time.Second
	}
	if since := time.Since(p.lastRun); since >= wait {
		wait = 0
	} else {
		wait -= since
	}
	p.schedule(wait)
}

// Stop stops the polling loop and cancels pending timers.
func (p *PollLoop) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopped = true
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
}

// Retime adjusts the timer according to the latest delay.
func (p *PollLoop) Retime() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	if p.timer != nil {
		p.timer.Stop()
	}
	wait := p.delay()
	if wait <= 0 {
		wait = time.Second
	}
	if since := time.Since(p.lastRun); since >= wait {
		wait = 0
	} else {
		wait -= since
	}
	p.schedule(wait)
}

func (p *PollLoop) schedule(wait time.Duration) {
	p.timer = time.AfterFunc(wait, func() {
		p.mu.Lock()
		if p.stopped {
			p.mu.Unlock()
			return
		}
		p.lastRun = time.Now()
		p.mu.Unlock()

		p.run()

		p.mu.Lock()
		defer p.mu.Unlock()
		if !p.stopped {
			p.schedule(p.delay())
		}
	})
}
