package later

import (
	"container/list"
	"time"
)

const (
	_defaultCap = 1000
)

type TimeWheel interface {
	Start()
	After(d time.Duration, fn CallbackFn)
	Every(d time.Duration, fn CallbackFn)
	Stop()
}

type timePanel struct {
	slots []*list.List
	num int
	idx int
	prev *timePanel
	next *timePanel
}

type CallbackFn func()

type task struct {
	rest []int
	callback CallbackFn
}

type addTask struct {
	d time.Duration
	callback CallbackFn
}

func (tp *timePanel) process() {
	tp.idx += 1
	if tp.idx == tp.num {
		tp.idx = 0
		if tp.next != nil {
			tp.next.process()
		}
	}
	idx := tp.idx
	l := tp.slots[idx]
	if l == nil {
		return
	}
	if l.Len() == 0 {
		return
	}
	// if this is the last, just run the expire job
	if tp.prev == nil {
		for l.Len() > 0 {
			elem := l.Remove(l.Front())
			if fn, ok := elem.(CallbackFn); ok {
				fn()
			}
		}
		return
	}
	// if not, propagate the task time
	prev := tp.prev
	for l.Len() > 0 {
		elem := l.Remove(l.Front())
		if job, ok := elem.(task); ok {
			n := len(job.rest) - 1
			idx := job.rest[n]
			if prev.slots[idx] == nil {
				prev.slots[idx] = list.New()
			}
			if n == 0 {
				prev.slots[idx].PushBack(job.callback)
			} else {
				job.rest = job.rest[:n]
				prev.slots[idx].PushBack(job)
			}
		}
	}
}

type timeWheel struct {
	panels []*timePanel
	interval time.Duration
	stop chan struct{}
	addChan chan addTask
}

// Stop terminates the running timingWheel
func (tw *timeWheel) Stop() {
	tw.stop <- struct{}{}
}

// NewSecondTimeWheel creates a timingWheel, the precision(interval) is one second.
// the max supported timeout is 24h-1s
func NewSecondTimeWheel() TimeWheel {
	return NewTimeWheel(time.Second, _defaultCap, 60, 60, 24)
}

// NewTimeWheel creates a timingWheel whose tick interval is $interval,
// cap is the size of channel that's used to store new job,
// slots for example: [60, 60, 24] means the first wheel has 60 slots, so is the second,
// and the last has 24 slots.
func NewTimeWheel(interval time.Duration, cap int, slots ...int) TimeWheel {
	tw := &timeWheel{
		panels:   nil,
		interval: interval,
		stop:     make(chan struct{}, 1),
		addChan:  make(chan addTask, cap),
	}
	for idx, slotNum := range slots {
		panel := &timePanel{
			slots: make([]*list.List, slotNum),
			num: slotNum,
			idx: 0,
		}
		tw.panels = append(tw.panels, panel)
		if idx != 0 {
			tw.panels[idx-1].next = panel
			panel.prev = tw.panels[idx-1]
		}
	}
	return tw
}

func (tw *timeWheel) register(d time.Duration, fn CallbackFn) {
	rest := tw.calcAccurateNextTime(d)
	n := len(rest)-1
	idx := rest[n]
	if tw.panels[n].slots[idx] == nil {
		tw.panels[n].slots[idx] = list.New()
	}
	if n > 0 {
		tw.panels[n].slots[idx].PushBack(task{
			rest: rest[:n],
			callback: fn,
		})
	} else {
		tw.panels[n].slots[idx].PushBack(fn)
	}
}

func (tw *timeWheel) calcAccurateNextTime(d time.Duration) []int {
	v := int(d / tw.interval)
	var res []int
	for _, tp := range tw.panels {
		acc := tp.idx + v
		m := acc % tp.num
		res = append(res, m)
		v /= tp.num
		if v == 0 {
			break
		}
	}
	return res
}

// Start starts the timingWheel
func (tw *timeWheel) Start() {
	if len(tw.panels) == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(tw.interval)
		defer ticker.Stop()
		for {
			select {
			case <-tw.stop:
				return
			case t := <-tw.addChan:
				tw.register(t.d, t.callback)
			case <-ticker.C:
				tw.panels[0].process()
			}
		}
	}()
}

// After runs the callback function after d
func (tw *timeWheel) After(d time.Duration, callback CallbackFn) {
	tw.addChan <- addTask{
		d:        d,
		callback: callback,
	}
}

// Every register a job which will be invoked every d,
// and for now there's no method to cancel one specific job
func (tw *timeWheel) Every(d time.Duration, callback CallbackFn) {
	fn := func() {
		callback()
		tw.Every(d, callback)
	}
	tw.After(d, fn)
}