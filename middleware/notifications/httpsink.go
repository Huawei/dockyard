package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type httpSink struct {
	url    string
	mu     sync.Mutex
	closed bool
	client *http.Client
}

func newHttpSink(u string, timeout time.Duration, headers http.Header) *httpSink {
	return &httpSink{
		url: u,
		client: &http.Client{
			Transport: &headerRoundTripper{
				Transport: http.DefaultTransport.(*http.Transport),
				headers:   headers,
			},
			Timeout: timeout,
		},
	}
}

func (hs *httpSink) Write(events ...Event) Error {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	defer hs.client.Transport.(*headerRoundTripper).CloseIdleConnections()

	if hs.closed {
		return Error{ErrSinkClosed, http.StatusInternalServerError}
	}

	envelope := Envelope{
		Events: events,
	}

	p, err := json.MarshalIndent(envelope, "", "   ")
	if err != nil {
		return Error{fmt.Errorf("%v: error marshaling event envelope: %v", hs, err), http.StatusBadRequest}
	}

	body := bytes.NewReader(p)
	resp, err := hs.client.Post(hs.url, EventsMediaType, body)
	if err != nil {
		return Error{fmt.Errorf("%v: error posting: %v", hs, err), http.StatusInternalServerError}
	}

	defer resp.Body.Close()

	return Error{nil, resp.StatusCode}
}

func (hs *httpSink) Close() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if hs.closed {
		return fmt.Errorf("httpsink: already closed")
	}

	hs.closed = true
	return nil
}

type headerRoundTripper struct {
	*http.Transport
	headers http.Header
}

func (hrt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var nreq http.Request
	nreq = *req
	nreq.Header = make(http.Header)

	merge := func(headers http.Header) {
		for k, v := range headers {
			nreq.Header[k] = append(nreq.Header[k], v...)
		}
	}

	merge(req.Header)
	merge(hrt.headers)

	return hrt.Transport.RoundTrip(&nreq)
}
