// package upyun is used for your UPYUN bucket
// this sdk implement purge api, form api, http rest api
package upyun

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "1.1.0"
)

// Auto: Auto detected, based on user's internet
// Telecom: (ISP) China Telecom
// Cnc:     (ISP) China Unicom
// Ctt:     (ISP) China Tietong
// purgeEndpoint: endpoint used for purging
// Default(Min/Max)ChunkSize: set the buffer size when doing copy operation
// defaultConnectTimeout: connection timeout when connect to upyun endpoint
const (
	Auto    = "v0.api.upyun.com"
	Telecom = "v1.api.upyun.com"
	Cnc     = "v2.api.upyun.com"
	Ctt     = "v3.api.upyun.com"

	purgeEndpoint = "purge.upyun.com"

	defaultChunkSize      = 32 * 1024
	defaultConnectTimeout = 60
)

// chunkSize: chunk size when copy
var (
	chunkSize = defaultChunkSize
	endpoints = [...]string{
		Auto, Telecom, Cnc, Ctt,
	}
)

// Util functions

// User Agent
func makeUserAgent() string {
	return fmt.Sprintf("UPYUN Go SDK %s", Version)
}

// Greenwich Mean Time
func genRFC1123Date() string {
	return time.Now().UTC().Format(time.RFC1123)
}

// make md5 from string
func md5Str(s string) (ret string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// URL encode
func encodeURL(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}

// Because of io.Copy use a 32Kb buffer, and, it is hard coded
// user can specify a chunksize with upyun.SetChunkSize
func chunkedCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, chunkSize)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])

			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return
}

// Use for http connection timeout
func timeoutDialer(timeout int) func(string, string) (net.Conn, error) {
	return func(network, addr string) (c net.Conn, err error) {
		c, err = net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
		if err != nil {
			return nil, err
		}
		return
	}
}

func SetChunkSize(chunksize int) {
	chunkSize = chunksize
}

type upYunHttpProxy struct {
	endpoint string

	httpClient *http.Client
}

// Set connect timeout
func (uhp *upYunHttpProxy) SetTimeout(timeout int) {
	uhp.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(timeout),
		},
	}
}

// Set http endpoint
func (uhp *upYunHttpProxy) SetEndpoint(endpoint string) (string, error) {
	for _, v := range endpoints {
		if v == endpoint {
			uhp.endpoint = endpoint
			return endpoint, nil
		}
	}

	err := fmt.Sprintf("Invalid endpoint, pick from Auto, Telecom, Cnc, Ctt")
	return uhp.endpoint, errors.New(err)
}

// FileInfo when use getlist
type Info struct {
	Size int64
	Time int64
	Name string
	Type string
}

func newInfo(s string) Info {
	infoList := strings.Split(s, "\t")
	size, _ := strconv.ParseInt(infoList[2], 10, 64)
	time, _ := strconv.ParseInt(infoList[3], 10, 64)

	return Info{
		Name: infoList[0],
		Type: infoList[1],
		Size: size,
		Time: time,
	}
}

// FileInfo when HEAD file
type FileInfo struct {
	Type string
	Date string
	Size int64
}

func newFileInfo(s string) (fileInfo FileInfo) {
	headers := strings.Split(s, "\n")
	for _, h := range headers {
		if h == "" {
			continue
		}

		tmp := strings.Split(h, ":")
		k, v := tmp[0], tmp[1]
		switch {
		case strings.Contains(k, "type"):
			fileInfo.Type = v
		case strings.Contains(k, "size"):
			fileInfo.Size, _ = strconv.ParseInt(v, 10, 64)
		case strings.Contains(k, "date"):
			fileInfo.Date = v
		}
	}

	return
}

// Request Error
type ReqError struct {
	err       error
	RequestId string
}

func newRespError(requestId string, respStatus string) *ReqError {
	return &ReqError{
		RequestId: requestId,
		err:       errors.New(respStatus),
	}
}

func (r *ReqError) Error() string {
	return r.err.Error()
}

// UPYUN HTTP FORM API
type UpYunForm struct {
	upYunHttpProxy

	httpClient *http.Client

	Key      string
	Bucket   string
	endpoint string
}

func NewUpYunForm(bucket, key string) *UpYunForm {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(defaultConnectTimeout),
		},
	}

	return &UpYunForm{
		Key:        key,
		Bucket:     bucket,
		endpoint:   Auto,
		httpClient: client,
	}
}

func (uf *UpYunForm) Put(saveas, path string, expireAfter int64,
	options map[string]string) error {
	if options == nil {
		options = make(map[string]string)
	}

	options["bucket"] = uf.Bucket
	options["save-key"] = saveas
	options["expiration"] = strconv.FormatInt(time.Now().Unix()+expireAfter, 10)

	args, err := json.Marshal(options)
	if err != nil {
		return err
	}

	policy := base64.StdEncoding.EncodeToString(args)
	sig := md5Str(policy + "&" + uf.Key)

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("policy", policy)
	writer.WriteField("signature", sig)
	part, err := writer.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return err
	}

	if _, err = chunkedCopy(part, file); err != nil {
		return err
	}

	writer.Close()

	url := fmt.Sprintf("http://%s/%s", uf.endpoint, uf.Bucket)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", makeUserAgent())

	resp, err := uf.httpClient.Do(req)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return errors.New(string(buf))
}

type UpYun struct {
	upYunHttpProxy

	httpClient *http.Client

	Bucket   string
	Username string
	Passwd   string

	Timeout   int
	ChunkSize int
}

func NewUpYun(bucket, username, passwd string) *UpYun {
	u := new(UpYun)
	u.Bucket = bucket
	u.Username = username
	u.Passwd = passwd

	u.Timeout = defaultConnectTimeout
	u.endpoint = Auto

	u.httpClient = &http.Client{}
	u.SetTimeout(u.Timeout)

	return u
}

func (u *UpYun) makeRestAuth(method, uri, date, lengthStr string) string {
	sign := []string{method, uri, date, lengthStr, md5Str(u.Passwd)}

	return "UpYun " + u.Username + ":" + md5Str(strings.Join(sign, "&"))
}

func (u *UpYun) makePurgeAuth(purgeList, date string) string {
	sign := []string{purgeList, u.Bucket, date, md5Str(u.Passwd)}

	return "UpYun " + u.Bucket + ":" + u.Username + ":" + md5Str(strings.Join(sign, "&"))
}

func (u *UpYun) Usage() (int64, error) {
	result, err := u.doRestRequest("GET", "/?usage", nil, nil)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result, 10, 64)
}

func (u *UpYun) Mkdir(key string) error {
	headers := make(map[string]string)

	headers["mkdir"] = "true"
	headers["folder"] = "true"

	_, err := u.doRestRequest("POST", key, headers, nil)

	return err
}

func (u *UpYun) Put(key string, value io.Reader, useMD5 bool, secret string) (string, error) {
	headers := make(map[string]string)
	headers["mkdir"] = "true"

	// secret

	if secret != "" {
		headers["Content-Secret"] = secret
	}

	// Get Content length

	/// if is file

	switch v := value.(type) {
	case *os.File:
		if useMD5 {
			hash := md5.New()

			_, err := chunkedCopy(hash, value)
			if err != nil {
				return "", err
			}

			headers["Content-MD5"] = fmt.Sprintf("%x", hash.Sum(nil))

			// seek to origin of file
			_, err = v.Seek(0, 0)
			if err != nil {
				return "", err
			}
		}

		fileInfo, err := v.Stat()
		if err != nil {
			return "", err
		}

		headers["Content-Length"] = strconv.FormatInt(fileInfo.Size(), 10)

		return u.doRestRequest("PUT", key, headers, value)
	case io.Reader:
		buf, err := ioutil.ReadAll(v)
		if err != nil {
			return "", err
		}

		headers["Content-Length"] = strconv.Itoa(len(buf))

		if useMD5 {
			headers["Content-MD5"] = fmt.Sprintf("%x", md5.Sum(buf))
		}

		return u.doRestRequest("PUT", key, headers, bytes.NewReader(buf))
	}

	return "", errors.New("Invalid Reader")
}

func (u *UpYun) Get(key string, value io.Writer) error {
	_, err := u.doRestRequest("GET", key, nil, value)

	return err
}

func (u *UpYun) Delete(key string) error {
	_, err := u.doRestRequest("DELETE", key, nil, nil)

	return err
}

func (u *UpYun) GetList(key string) ([]Info, error) {
	ret, err := u.doRestRequest("GET", key, nil, nil)
	if err != nil {
		return nil, err
	}

	list := strings.Split(ret, "\n")
	infos := make([]Info, len(list))

	for i, v := range list {
		infos[i] = newInfo(v)
	}

	return infos, nil
}

func (u *UpYun) GetInfo(key string) (FileInfo, error) {
	ret, err := u.doRestRequest("HEAD", key, nil, nil)
	if err != nil {
		return FileInfo{}, err
	}

	fileInfo := newFileInfo(ret)

	return fileInfo, nil
}

func (u *UpYun) Purge(urls []string) (string, error) {
	purge := fmt.Sprintf("http://%s/purge/", purgeEndpoint)

	date := genRFC1123Date()
	purgeList := strings.Join(urls, "\n")

	headers := make(map[string]string)
	headers["Date"] = date
	headers["Authorization"] = u.makePurgeAuth(purgeList, date)
	headers["Content-Type"] = "application/x-www-form-urlencoded;charset=utf-8"

	form := make(url.Values)
	form.Add("purge", purgeList)

	body := strings.NewReader(form.Encode())
	resp, err := u.doHttpRequest("POST", purge, headers, body)
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 == 2 {
		result := make(map[string][]string)
		json.Unmarshal(content, result)

		return strings.Join(result["invalid_domain_of_url"], ","), nil
	}

	return "", errors.New(string(content))
}

func (u *UpYun) doRestRequest(method, uri string, headers map[string]string,
	value interface{}) (result string, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	// Normalize url
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	uri = "/" + u.Bucket + uri

	url := fmt.Sprintf("http://%s%s", u.endpoint, uri)

	// date
	date := genRFC1123Date()

	// auth
	lengthStr, ok := headers["Content-Length"]
	if !ok {
		lengthStr = "0"
	}

	headers["Date"] = date
	headers["Authorization"] = u.makeRestAuth(method, uri, date, lengthStr)

	// Get method
	rc, ok := value.(io.Reader)
	if !ok {
		rc = nil
	}

	resp, err := u.doHttpRequest(method, url, headers, rc)
	if err != nil {
		return "", err
	}

	if _, ok := value.(io.Closer); ok {
		defer resp.Body.Close()
	}

	// retrive request id
	requestId := "Unknown"

	requestIds, ok := resp.Header[http.CanonicalHeaderKey("X-Request-Id")]
	if ok {
		requestId = strings.Join(requestIds, ",")
	}

	if (resp.StatusCode / 100) == 2 {
		if method == "GET" && value != nil {
			written, err := chunkedCopy(value.(io.Writer), resp.Body)

			return strconv.FormatInt(written, 10), err
		} else if method == "GET" && value == nil {
			body, err := ioutil.ReadAll(resp.Body)
			return string(body[:]), err
		} else if method == "PUT" || method == "HEAD" {
			var headerStrings string
			for k, v := range resp.Header {
				if strings.Contains(strings.ToLower(k), "x-upyun-") {
					headerStrings += fmt.Sprintf("%s:%s\n", strings.ToLower(k), v[0])
				}
			}
			return headerStrings, nil
		} else {
			return "", nil
		}
	}

	return "", newRespError(requestId, resp.Status)
}

func (u *UpYun) doHttpRequest(method, url string, headers map[string]string,
	body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// User Agent
	req.Header.Set("User-Agent", makeUserAgent())

	// https://code.google.com/p/go/issues/detail?id=6738
	if method == "PUT" || method == "POST" {
		length := req.Header.Get("Content-Length")
		req.ContentLength, _ = strconv.ParseInt(length, 10, 64)
	}

	return u.httpClient.Do(req)
}
