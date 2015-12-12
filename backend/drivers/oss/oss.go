package oss

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

func init() {
	drivers.Register("oss", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("osssave", osssave)
}

// Call func for http request
func call(method, baseUrl, path string, body io.Reader, headers map[string][]string) ([]byte, int, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, baseUrl+path, body)
	if err != nil {
		return nil, 408, err
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			return nil, resp.StatusCode, err
		}
		return nil, http.StatusNotFound, err
	}

	dataBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return dataBody, resp.StatusCode, nil
}

func setHeader(header map[string][]string) {

	path := "repositories/username/ubuntu/tag_v2"
	header["Path"] = []string{path}
	header["Fragment-Index"] = []string{"0"}
	header["Bytes-Range"] = []string{"0-19"}
	header["Is-Last"] = []string{"true"}
}

func osssave(file string) (url string, err error) {
	s := "hello world content"
	header := make(map[string][]string, 0)
	setHeader(header)

	// Reorgnization of ret url when err occurs
	var key string
	// Get the filename from the file , eg,get "1.txt" from /home/liugenping/1.txt
	for _, key = range strings.Split(file, "/") {

	}
	opath := "/" + setting.Bucket + "/" + key
	url = "http://" + setting.Endpoint + opath

	filep, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer filep.Close()

	result, statusCode, err := call("POST", "http://127.0.0.1:6788", "/v1/file", filep, header)
	if statusCode != http.StatusOK {
		err = fmt.Errorf("statusCode error: %d", statusCode, ", error: ", err)
	}

	if nil != err {
		return "", err
	} else {
		return url, nil
	}
}
