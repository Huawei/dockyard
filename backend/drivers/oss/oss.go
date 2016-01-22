package oss

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

const (
	APIaddress = "0.0.0.0"
)

type Fileinfo struct {
	Index   int    `json:"index"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Islast  bool   `json:"isLast"`
	Modtime string `json:"modTime"`
}

func init() {
	drivers.Register("oss", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("osssave", osssave)
}

func osssave(filepath string) error {
	//TODO: define the naming rules of path
	path := filepath
	partSize := setting.PartSizeMB * 1024 * 1024

	//calculate the fragment number according to settings
	file, err := os.Open(filepath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("open file %s err: %v \n", filepath, err)
	}
	fileinfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get file %s info err: %v \n", filepath, err)
	}
	fileSize := fileinfo.Size()
	fileBody := make([]byte, fileSize)
	nread, err := file.Read(fileBody)
	if err != nil || int64(nread) != fileSize {
		return fmt.Errorf("[oss save] read %s nread: %v, fileSize: %v, err: %v", filepath, nread, fileSize, err)
	}
	partCount := int(fileSize / int64(partSize))
	partial := int(fileSize % int64(partSize))

	if partCount == 0 && partial == 0 {
		return nil
	}

	// if data divide into only one fragment
	if partCount == 0 && partial != 0 {
		err := postFile(path, fileBody[0:partial], 0, 0, int64(partial), true)
		if err != nil {
			return fmt.Errorf("oss save file error: %v", err)
		}
		return nil
	}

	//if data divide into more than one fragment
	var wg sync.WaitGroup
	wg.Add(partCount)
	result := make([]int, partCount)
	begin := 0
	end := 0
	for k := 0; k < partCount; k++ {
		go func(k int) {
			defer wg.Done()
			begin = k * partSize
			end = (k + 1) * partSize
			err := postFile(path, fileBody[begin:end], k, int64(begin), int64(end), false)
			if err != nil {
				fmt.Errorf("oss save file error: %v \n", err)
			}
			result[k] = 1
		}(k)
		runtime.Gosched()
	}
	wg.Wait()
	k := partCount
	begin = k * partSize
	if partial != 0 {
		end = begin + partial
	}
	err = postFile(path, fileBody[begin:end], k, int64(begin), int64(end), true)
	if err != nil {
		return fmt.Errorf("oss save file error: %v", err)
	}
	//check if all fragments successfully saved
	for i := 0; i < partCount; i++ {
		if result[i] != 1 {
			return fmt.Errorf("oss save file error, fragment %d error", i)
		}
	}
	fmt.Printf("oss save file %v finish", filepath)
	return nil
}

func ossgetfileinfo(filepath string) error {
	var apiserveraddr string
	switch setting.ListenMode {
	case "https":
		apiserveraddr = fmt.Sprintf("https://%v:%v", APIaddress, setting.APIHttpsPort)
	default:
		apiserveraddr = fmt.Sprintf("http://%v:%v", APIaddress, setting.APIPort)
	}
	header := make(map[string][]string, 0)
	header["Path"] = []string{filepath}
	result, statusCode, err := call("GET", apiserveraddr, "/oss/api/file/info", nil, header)
	if statusCode != http.StatusOK {
		return fmt.Errorf("statusCode error: %d", statusCode, ", error: ", err)
	}
	if err != nil {
		return fmt.Errorf("error: ", err)
	}
	fmt.Printf("fileinfo: %s\n", string(result))
	return nil
}

func ossdownload(tag string, path string) error {
	var apiserveraddr string
	switch setting.ListenMode {
	case "https":
		apiserveraddr = fmt.Sprintf("https://%v:%v", APIaddress, setting.APIHttpsPort)
	default:
		apiserveraddr = fmt.Sprintf("http://%v:%v", APIaddress, setting.APIPort)
	}
	// get file information
	header := make(map[string][]string, 0)
	header["Path"] = []string{tag}
	result, statusCode, err := call("GET", apiserveraddr, "/oss/api/file/info", nil, header)
	if statusCode != http.StatusOK {
		return fmt.Errorf("statusCode error: %d", statusCode, ", error: ", err)
	}
	result = bytes.TrimPrefix(result, []byte("{\"fragment-info\":"))
	result = bytes.TrimSuffix(result, []byte("}"))

	// tranform fileinfo data from json format to Fileinfo struct
	var files []Fileinfo
	json.Unmarshal([]byte(result), &files)
	fragNum := len(files)
	data_frag := make(map[int][]byte)

	for _, file := range files {
		// transform fileinfo to header
		headerfile := make(map[string][]string, 0)
		headerfile["Path"] = []string{tag}
		index := fmt.Sprintf("%v", file.Index)
		headerfile["Fragment-Index"] = []string{index}
		fragrange := fmt.Sprintf("%v-%v", file.Start, file.End)
		headerfile["Bytes-Range"] = []string{fragrange}
		islast := fmt.Sprintf("%v", file.Islast)
		headerfile["Is-Last"] = []string{islast}
		// sent http request and get data
		data, statusCode, err := call("GET", apiserveraddr, "/oss/api/file", nil, headerfile)
		if statusCode != http.StatusOK {
			return fmt.Errorf("statusCode error: %d", statusCode, ", error: ", err)
		}
		data_frag[file.Index] = data
	}

	// write data into local file
	localfile, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModeType)
	if err != nil {
		return fmt.Errorf("open local file error:%v", err.Error())
	}
	defer localfile.Close()
	var data []byte
	for i := 0; i < fragNum; i++ {
		data = append(data, data_frag[i]...)
	}
	if err = ioutil.WriteFile(path, data, 0777); err != nil {
		return fmt.Errorf("write local file error:%v", err.Error())
	}

	//md5 generate
	md5 := md5.Sum(data)
	fmt.Printf("md5=%x \n", md5)
	return nil
}

func ossdel(filepath string) error {
	var apiserveraddr string
	switch setting.ListenMode {
	case "https":
		apiserveraddr = fmt.Sprintf("https://%v:%v", APIaddress, setting.APIHttpsPort)
	default:
		apiserveraddr = fmt.Sprintf("http://%v:%v", APIaddress, setting.APIPort)
	}
	header := make(map[string][]string, 0)
	header["Path"] = []string{filepath}

	_, statusCode, err := call("DELETE", apiserveraddr, "/oss/api/file", nil, header)

	if statusCode != http.StatusNoContent {
		return fmt.Errorf("statusCode error: %d", statusCode, ", error: ", err)
	}
	if err != nil {
		return fmt.Errorf("error: ", err)
	}
	return nil
}

// Call func for http request
func call(method, baseUrl, path string, body io.Reader, headers map[string][]string) ([]byte, int, error) {
	client := &http.Client{}
	switch setting.ListenMode {
	case "https":
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	default:
	}

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

func postFile(path string, data []byte, index int, begin int64, end int64, isLast bool) error {
	var apiserveraddr string
	switch setting.ListenMode {
	case "https":
		apiserveraddr = fmt.Sprintf("https://%v:%v", APIaddress, setting.APIHttpsPort)
	default:
		apiserveraddr = fmt.Sprintf("http://%v:%v", APIaddress, setting.APIPort)
	}
	header := make(map[string][]string)
	header["Path"] = []string{path}
	header["Fragment-Index"] = []string{fmt.Sprintf("%v", index)}
	header["Bytes-Range"] = []string{fmt.Sprintf("%v-%v", begin, end)}
	header["Is-Last"] = []string{fmt.Sprintf("%v", isLast)}

	_, statusCode, err := call("POST", apiserveraddr, "/oss/api/file", bytes.NewBuffer(data), header)
	if err != nil || statusCode != http.StatusOK {
		return fmt.Errorf("[postFile] failed, path: %s, error: %v, statusCode: %v", path, err, statusCode)
	}
	return nil
}
