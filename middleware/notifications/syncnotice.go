package notifications

import (
	"fmt"
)

type syncNotice struct {
	sinks []Sink
}

func newSyncNotice(sinks ...Sink) *syncNotice {
	return &syncNotice{sinks: sinks}
}

func (b *syncNotice) Close() error {
	for _, sink := range b.sinks {
		if err := sink.Close(); err != nil {
			fmt.Printf("syncNotice: fail to close sink %v, err: %v", sink, err)
		}
	}

	return nil
}

func (b *syncNotice) Write(events ...Event) (finalErr Error) {
	num := len(b.sinks)
	closed := make(chan Error, num)
	finalErr = Error{nil, 200}

	writeFunc := func(sink Sink) {
		err := sink.Write(events...)
		if err.Err != nil {
			fmt.Printf("syncNotice: error writing events %v, err: %v", sink, err.Err)
		}

		closed <- err
	}

	for _, sink := range b.sinks {
		go writeFunc(sink)
	}

	for i := 0; i < num; i++ {
		err := <-closed
		if err.Err != nil || err.StatusCode >= 300 {
			finalErr = err
		}
	}
	return
}
