package upyun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// UPYUN MEDIA API
type UpYunMedia struct {
	upYunHTTPCore // HTTP Core

	Username string
	Passwd   string
	Bucket   string
}

// status response
type MediaStatusResp struct {
	Tasks map[string]interface{} `json:"tasks"`
}

// NewUpYunMedia returns a new UPYUN Media API client given
// a bucket name, username, password. http client connection
// timeout is set to defalutConnectionTimeout which
// is equal to 60 seconds.

func NewUpYunMedia(bucket, user, pass string) *UpYunMedia {
	up := &UpYunMedia{
		Username: user,
		Passwd:   md5Str(pass),
		Bucket:   bucket,
	}

	client := &http.Client{}
	up.SetTimeout(defaultConnectTimeout)

	up.endpoint = "p0.api.upyun.com"
	up.httpClient = client

	return up
}

func (upm *UpYunMedia) makeMediaAuth(kwargs map[string]string) string {
	var keys []string
	for k, _ := range kwargs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	auth := ""
	for _, k := range keys {
		auth += k + kwargs[k]
	}

	return fmt.Sprintf("UPYUN %s:%s", upm.Username,
		md5Str(upm.Username+auth+upm.Passwd))
}

// Send Media Tasks Reqeust
func (upm *UpYunMedia) PostTasks(src, notify, accept string,
	tasks []map[string]interface{}) ([]string, error) {
	data, err := json.Marshal(tasks)
	if err != nil {
		return nil, err
	}

	kwargs := map[string]string{
		"bucket_name": upm.Bucket,
		"source":      src,
		"notify_url":  notify,
		"tasks":       base64Str(data),
		"accept":      accept,
	}

	resp, err := upm.doMediaRequest("POST", "/pretreatment", kwargs)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/2 == 100 {
		var ids []string
		err = json.Unmarshal(buf, &ids)
		if err != nil {
			return nil, err
		}
		return ids, err
	}

	return nil, errors.New(string(buf))
}

// Get Task Progress
func (upm *UpYunMedia) GetProgress(task_ids string) (*MediaStatusResp, error) {

	kwargs := map[string]string{
		"bucket_name": upm.Bucket,
		"task_ids":    task_ids,
	}

	resp, err := upm.doMediaRequest("GET", "/status", kwargs)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode/2 == 100 {
		var status MediaStatusResp
		if err := json.Unmarshal(buf, &status); err != nil {
			return nil, err
		}
		return &status, nil
	}

	return nil, errors.New(string(buf))
}

func (upm *UpYunMedia) doMediaRequest(method, path string,
	kwargs map[string]string) (*http.Response, error) {

	// Normalize url
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("http://%s%s", upm.endpoint, path)

	// Set Headers
	headers := make(map[string]string)
	date := genRFC1123Date()
	headers["Date"] = date
	headers["Authorization"] = upm.makeMediaAuth(kwargs)

	// Payload
	var options []string
	for k, v := range kwargs {
		options = append(options, k+"="+v)
	}
	payload := strings.Join(options, "&")

	if method == "GET" {
		url = url + "?" + payload
		return upm.doHTTPRequest(method, url, headers, nil)
	} else {
		if method == "POST" {
			headers["Content-Length"] = fmt.Sprint(len(payload))
			return upm.doHTTPRequest(method, url, headers,
				strings.NewReader(payload))
		}
	}

	return nil, errors.New("Unknown method")
}
