package notifications

import (
	"net/http"
	"time"
)

type EndpointDesc struct {
	Name      string
	URL       string
	Headers   http.Header
	Timeout   time.Duration
	Threshold int
	Backoff   time.Duration
	EventDB   string
	Disabled  bool
}

type Endpoint struct {
	Sink
	EndpointDesc
}

func newEndpoint(e EndpointDesc) *Endpoint {
	var endpoint Endpoint

	endpoint.EndpointDesc = e

	endpoint.defaults()
	endpoint.Sink = newHttpSink(endpoint.URL, endpoint.Timeout*time.Millisecond, endpoint.Headers)
	endpoint.Sink = newRetryingSink(endpoint.Sink, endpoint.Threshold, endpoint.Backoff*time.Millisecond)

	return &endpoint
}

func (e *EndpointDesc) defaults() {
	if e.Timeout <= 0 {
		e.Timeout = 3 * time.Second
	}

	if e.Threshold <= 0 {
		e.Threshold = 10
	}

	if e.Backoff <= 0 {
		e.Backoff = 5 * time.Second
	}
}
