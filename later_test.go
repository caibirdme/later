package later

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"time"
)

func TestTimeWheel_After(t *testing.T) {
	should := require.New(t)
	tw := NewSecondTimeWheel()
	tw.Start()
	//time.Sleep(10*time.Millisecond)
	start := time.Now()
	delay := 2*time.Second
	finish := make(chan struct{}, 1)
	tw.After(delay, func() {
		finish <- struct{}{}
	})
	<-finish
	now := time.Now()
	diffT := now.Sub(start)
	diff := math.Abs(diffT.Seconds() - delay.Seconds())
	should.True(diff <= 1e-3, "diff: %f, start: %s, now: %s", diff, start, now)
	var t1, t2 time.Time
	tw.After(3*time.Second, func() {
		t1 = time.Now()
		finish <- struct{}{}
	})
	tw.After(time.Second, func() {
		t2 = time.Now()
	})
	<-finish
	diffT = t1.Sub(t2)
	diff = math.Abs(diffT.Seconds() - (2*time.Second).Seconds())
	should.True(diff <= 1e-3)
}

func TestHierarchicalWheel(t *testing.T) {
	tw := NewTimeWheel(50*time.Millisecond, _defaultCap, 20, 60, 60)
	tw.Start()
	should := require.New(t)
	start := time.Now()
	finish := make(chan struct{})
	var now time.Time
	delay := 60*time.Second
	tw.After(delay, func() {
		now = time.Now()
		finish <- struct{}{}
	})
	<-finish
	diffT := now.Sub(start)
	diff := math.Abs(diffT.Seconds() - delay.Seconds())
	should.True(diff <= 1e-3, "diff: %f, start: %s, now: %s", diff, start, now)
}

func TestTimeWheel_Stop(t *testing.T) {
	should := require.New(t)
	tw := NewSecondTimeWheel()
	tw.Start()
	finish := make(chan struct{})
	tw.After(2*time.Second, func() {
		finish <- struct{}{}
	})
	tw.Stop()
	// finish will never get invoked
	// let's wait for 3s to see
	time.Sleep(3*time.Second)
	select {
	case <-finish:
		should.FailNow("cancel fail")
	default:
	}
}

func TestTimeWheel_Every(t *testing.T) {
	should := require.New(t)
	tw := NewSecondTimeWheel()
	tw.Start()
	count := 0
	tw.Every(time.Second, func() {
		count ++
	})
	time.Sleep(5*time.Second+300*time.Millisecond)
	tw.Stop()
	should.Equal(5, count)
}