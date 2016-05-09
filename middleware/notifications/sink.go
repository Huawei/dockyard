package notifications

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type retryingSink struct {
	mu     sync.Mutex
	sink   Sink
	closed bool

	failures struct {
		threshold int
		recent    int
		last      time.Time
		backoff   time.Duration
	}
}

func newRetryingSink(sink Sink, threshold int, backoff time.Duration) *retryingSink {
	rs := &retryingSink{
		sink: sink,
	}
	rs.failures.threshold = threshold
	rs.failures.backoff = backoff

	return rs
}

func (rs *retryingSink) Write(events ...Event) Error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

retry:
	if rs.closed {
		return Error{ErrSinkClosed, http.StatusInternalServerError}
	}

	if !rs.proceed() {
		return Error{fmt.Errorf("%v encountered too many errors, return", rs.sink), http.StatusInternalServerError}
	}

	err := rs.write(events...)
	if err.Err != nil {
		if err.Err == ErrSinkClosed {
			return Error{ErrSinkClosed, http.StatusInternalServerError}
		}

		fmt.Printf("retryingsink: error writing events: %v, retrying", err.Err)
		rs.wait(rs.failures.backoff)
		goto retry
	}

	return err
}

func (rs *retryingSink) Close() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.closed {
		return fmt.Errorf("retryingsink: already closed")
	}

	rs.closed = true
	return rs.sink.Close()
}

func (rs *retryingSink) write(events ...Event) Error {
	err := rs.sink.Write(events...)
	if err.Err != nil {
		rs.failure()
		return err
	}
	rs.reset()

	return err
}

func (rs *retryingSink) wait(backoff time.Duration) {
	rs.mu.Unlock()
	defer rs.mu.Lock()

	time.Sleep(backoff)
}

func (rs *retryingSink) reset() {
	rs.failures.recent = 0
	rs.failures.last = time.Time{}
}

func (rs *retryingSink) failure() {
	rs.failures.recent++
	rs.failures.last = time.Now().UTC()
}

func (rs *retryingSink) proceed() bool {
	return rs.failures.recent < rs.failures.threshold
}
