package upyun

import (
	"bytes"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

// Auto: Auto detected, based on user's internet
// Telecom: (ISP) China Telecom
// Cnc:     (ISP) China Unicom
// Ctt:     (ISP) China Tietong
const (
	Auto = iota
	Telecom
	Cnc
	Ctt
)

type upYunHTTPCore struct {
	endpoint   string
	httpClient *http.Client
}

func (core *upYunHTTPCore) SetTimeout(timeout int) {
	core.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (c net.Conn, err error) {
				c, err = net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
				if err != nil {
					return nil, err
				}
				return
			},
			// http://studygolang.com/articles/3138
			// DisableKeepAlives: true,
		},
	}
}

// do http form request
func (core *upYunHTTPCore) doFormRequest(url, policy, sign,
	fname string, fd io.Reader) (*http.Response, error) {

	body := &bytes.Buffer{}
	headers := make(map[string]string)

	// generate form data
	err := func() error {
		writer := multipart.NewWriter(body)
		defer writer.Close()

		var err error
		var part io.Writer

		writer.WriteField("policy", policy)
		writer.WriteField("signature", sign)
		if part, err = writer.CreateFormFile("file", filepath.Base(fname)); err == nil {
			if _, err = chunkedCopy(part, fd); err == nil {
				headers["Content-Type"] = writer.FormDataContentType()
			}
		}
		return err
	}()

	if err != nil {
		return nil, err
	}

	return core.doHTTPRequest("POST", url, headers, body)
}

// do http request
func (core *upYunHTTPCore) doHTTPRequest(method, url string, headers map[string]string,
	body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// User Agent
	req.Header.Set("User-Agent", makeUserAgent())

	// https://code.google.com/p/go/issues/detail?id=6738
	if method == "PUT" || method == "POST" {
		length := req.Header.Get("Content-Length")
		if length != "" {
			req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
		}
	}

	return core.httpClient.Do(req)
}
