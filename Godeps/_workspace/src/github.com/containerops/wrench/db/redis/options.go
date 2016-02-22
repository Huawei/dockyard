package redis

import (
	"time"
)

const minWaitRetry = 10 * time.Millisecond

type LockOptions struct {
	// The maximum duration to lock a key for
	// Default: 5s
	LockTimeout time.Duration

	// The maximum amount of time you are willing to wait to obtain that lock
	// Default: 0 = do not wait
	WaitTimeout time.Duration

	// In case WaitTimeout is activated, this it the amount of time you are willing
	// to wait between retries.
	// Default: 100ms, must be at least 10ms
	WaitRetry time.Duration
}

func (o *LockOptions) normalize() *LockOptions {
	if o == nil {
		o = new(LockOptions)
	}
	if o.LockTimeout < 1 {
		o.LockTimeout = 5 * time.Second
	}
	if o.WaitTimeout < 0 {
		o.WaitTimeout = 0
	}
	if o.WaitRetry < minWaitRetry {
		o.WaitRetry = minWaitRetry
	}
	return o
}
