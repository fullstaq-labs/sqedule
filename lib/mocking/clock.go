package mocking

import "time"

// IClock is a swappable clock interface so that during testing
// the time can be mocked.
type IClock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

// RealClock is an IClock which returns the real time.
type RealClock struct{}

func (_ RealClock) Now() time.Time {
	return time.Now()
}

func (_ RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

// FakeClock is an IClock which returns the embedded time value
// instead of the real time.
type FakeClock struct {
	// Value is the embedded time value. This value is advanced
	// by calling Sleep().
	Value time.Time
}

func (c FakeClock) Now() time.Time {
	return c.Value
}

func (c *FakeClock) Sleep(d time.Duration) {
	c.Value.Add(d)
}
