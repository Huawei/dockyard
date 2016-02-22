package qcloud

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/containerops/dockyard/backend/driver"
	"github.com/containerops/wrench/setting"
)

func init() {
	driver.Register("qcloud", InitFunc)
}

func InitFunc() {
	driver.InjectReflect.Bind("qcloudsave", qcloudsave)
}

//according to data format of qcloud restful api
func makePlainText(params map[string]interface{}) (plainText string) {

	var plainParms string
	plainParms += "&" + "a" + "=" + fmt.Sprintf("%v", params["a"])
	plainParms += "&" + "b" + "=" + fmt.Sprintf("%v", params["b"])
	plainParms += "&" + "k" + "=" + fmt.Sprintf("%v", params["k"])
	plainParms += "&" + "e" + "=" + fmt.Sprintf("%v", params["e"])
	plainParms += "&" + "t" + "=" + fmt.Sprintf("%v", params["t"])
	plainParms += "&" + "r" + "=" + fmt.Sprintf("%v", params["r"])
	plainParms += "&" + "f" + "=" + fmt.Sprintf("%v", params["f"])
	plainText = plainParms[1:]

	return plainText
}

//generate a signature according to qcloud restful api
func Sign(plainText string, secretKey string) (sign string) {
	hmacObj := hmac.New(sha1.New, []byte(secretKey))
	hmacObj.Write([]byte(plainText))
	signObj := string(hmacObj.Sum(nil)) + plainText
	sign = base64.StdEncoding.EncodeToString([]byte(signObj))
	return
}

func qcloudsave(file string) (url string, err error) {
	var key string
	//get the filename from the file , eg,get "1.txt" from /home/liugenping/1.txt
	for _, key = range strings.Split(file, "/") {

	}
	fin, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fin.Close()
	fileName := key
	requestUrl := "http://" + setting.Endpoint + "/" + "files" + "/" + "v1" + "/" + setting.QcloudAccessID + "/" + setting.Bucket + "/" + fileName
	url, err = GetRequest(fin, file, requestUrl)
	return url, err
}

func GenerateSign() (sign string) {

	params := map[string]interface{}{}
	params["a"] = setting.QcloudAccessID
	params["b"] = setting.Bucket
	params["k"] = setting.AccessKeyID
	time_bg := fmt.Sprintf("%v", time.Now().Unix())
	time_en := fmt.Sprintf("%v", time.Now().Unix()+2000000)
	params["e"] = time_en
	params["t"] = time_bg
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	rands := fmt.Sprintf("%v", rand.Intn(100))
	params["r"] = rands
	params["f"] = ""

	uploadPlainText := makePlainText(params)
	sign = Sign(uploadPlainText, setting.AccessKeysecret)

	return
}

var Criticalsize int64 = 2 * 1024 * 1024

func upload_prepare(fin *os.File, filesize int64, requestUrl string) (session string, err error) {

	h := sha1.New()
	_, err = io.Copy(h, fin)
	if err != nil {
		return "", err
	}
	extraparams := map[string]string{
		"op":         "upload_slice",
		"filesize":   fmt.Sprintf("%v", filesize),
		"sha":        fmt.Sprintf("%x", h.Sum(nil)),
		"slice_size": fmt.Sprintf("%v", Criticalsize),
	}
	h = sha1.New()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range extraparams {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", requestUrl, body)
	if err != nil {
		return "", err
	}

	header := make(http.Header)
	header.Set("Content-Type", writer.FormDataContentType())
	header.Set("Authorization", GenerateSign())
	req.Header = header

	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	out := []byte{}
	out, err = ioutil.ReadAll(rsp.Body)
	type drsp struct {
		Offset     int64  `json:"offset"`
		Session    string `json:"session"`
		Slice_size int    `json:"slice_size"`
		Url        string `json:"url"`
	}
	type qrsp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    drsp   `json:"data"`
	}
	Qrsp := qrsp{}
	err = json.Unmarshal(out, &Qrsp)
	if Qrsp.Data.Url != "" {
		return Qrsp.Data.Url, errors.New("file exists")
	}
	session = Qrsp.Data.Session

	return session, nil
}

func upload_follow(fin *os.File, session string, filesize int64, requestUrl string) (url string, err error) {

	type drsp struct {
		Offset     int64  `json:"offset"`
		Session    string `json:"session"`
		Slice_size int    `json:"slice_size"`
		Access_url string `json:"access_url"`
	}
	type qrsp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    drsp   `json:"data"`
	}
	var offsetsta int64
	var offsetend int64
	var cnt int
	var uploaddata = make([]byte, Criticalsize)
	var header = make(http.Header)

	for {
		var filedata string
		offsetsta = offsetend
		offsetend += Criticalsize
		if offsetend >= filesize {
			offsetend = filesize
			uploadend := make([]byte, offsetend-offsetsta)
			_, err = fin.ReadAt(uploadend, offsetsta)
			filedata = string(uploadend)
		} else {
			_, err = fin.ReadAt(uploaddata, offsetsta)
			filedata = string(uploaddata)
		}
		extraparams := map[string]string{
			"op":          "upload_slice",
			"filecontent": filedata,
			"sha":         fmt.Sprintf("%x", sha1.Sum([]byte(filedata))),
			"session":     session,
			"offset":      fmt.Sprintf("%v", offsetsta),
		}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for key, val := range extraparams {
			_ = writer.WriteField(key, val)
		}
		err = writer.Close()
		if err != nil {
			return "", err
		}

		header.Set("Content-Type", writer.FormDataContentType())
		header.Set("Authorization", GenerateSign())
		for cnt = 0; ; cnt++ {
			req, err := http.NewRequest("POST", requestUrl, body)
			if err != nil {
				return "", err
			}
			req.Header = header
			client := &http.Client{}

			rsp, err := client.Do(req)
			if err != nil {
				return "", err
			}
			out, err := ioutil.ReadAll(rsp.Body)
			Qrsp := qrsp{}
			err = json.Unmarshal(out, &Qrsp)
			session = Qrsp.Data.Session
			url = Qrsp.Data.Access_url
			if Qrsp.Code == 0 {
				break
			}
			if cnt == 3 {
				return "The net is disconnected!!!", err
			}
		}
		if offsetend >= filesize {
			break
		}
	}
	return url, err
}

func upload_slice(fin *os.File, file string, requestUrl string) (url string, err error) {

	fi, err := fin.Stat()
	if err != nil {
		return "", err
	}
	filesize := fi.Size()

	session, err := upload_prepare(fin, filesize, requestUrl)
	if err != nil {
		return session, err
	}
	url, err = upload_follow(fin, session, filesize, requestUrl)
	return url, err
}

func GetRequest(fin *os.File, file string, requestUrl string) (url string, err error) {

	fi, err := fin.Stat()
	if err != nil {
		return "", err
	}
	filesize := fi.Size()

	if filesize > Criticalsize {
		url, err = upload_slice(fin, file, requestUrl)
	} else {
		h := sha1.New()
		_, err = io.Copy(h, fin)
		if err != nil {
			return "", err
		}
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		filedata := string(data)
		extraparams := map[string]string{
			"op":          "upload",
			"filecontent": filedata,
			"sha":         fmt.Sprintf("%x", h.Sum(nil)),
		}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for key, val := range extraparams {
			_ = writer.WriteField(key, val)
		}
		err = writer.Close()
		if err != nil {
			return "", err
		}

		req, err := http.NewRequest("POST", requestUrl, body)
		if err != nil {
			return "", err
		}

		header := make(http.Header)
		header.Set("Content-Type", writer.FormDataContentType())
		header.Set("Authorization", GenerateSign())
		req.Header = header
		url, err = CliDo(req)
	}
	return
}

func CliDo(req *http.Request) (url string, err error) {
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	out := []byte{}
	out, err = ioutil.ReadAll(rsp.Body)
	type drsp struct {
		Access_url string `json:"access_url"`
	}
	type qrsp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    drsp   `json:"data"`
	}
	Qrsp := qrsp{}
	err = json.Unmarshal(out, &Qrsp)
	if Qrsp.Code == 0 {
		url = Qrsp.Data.Access_url
	}
	if Qrsp.Code == -4018 {
		return "file exists", errors.New("file exists")
	}
	return url, nil
}
