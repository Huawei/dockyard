package backend

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego/config"
)

var (
	QcloudEndpoint        string
	QcloudAccessID        string
	QcloudBucket          string
	QcloudAccessKeyID     string
	QcloudAccessKeySecret string
)

func init() {
	fmt.Println("qcloud")
	InjectReflect.Bind("qcloudsave", qcloudsave)
}

func qcloudSetconfig(conf config.ConfigContainer) error {
	QcloudEndpoint = conf.String("qcloud::endpoint")
	if QcloudEndpoint == "" {
		return fmt.Errorf("Read endpoint of qcloud failed!")
	}

	QcloudAccessID = conf.String("qcloud::accessID")
	if QcloudAccessID == "" {
		return fmt.Errorf("Read accessID of qcloud failed!")
	}

	QcloudBucket = conf.String("qcloud::bucket")
	if QcloudBucket == "" {
		return fmt.Errorf("Read bucket qcloud failed!")
	}

	QcloudAccessKeyID = conf.String("qcloud::accessKeyID")
	if QcloudAccessKeyID == "" {
		return fmt.Errorf("Read accessKeyID of qcloud failed!")
	}

	QcloudAccessKeySecret = conf.String("qcloud::accessKeysecret")
	if QcloudAccessKeySecret == "" {
		return fmt.Errorf("Read accessKeysecret of qcloud failed!")
	}
	return nil
}

func makePlainText(api string, params map[string]interface{}) (plainText string) {
	// sort
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var plainParms string
	for i := range keys {
		k := keys[i]
		plainParms += "&" + fmt.Sprintf("%v", k) + "=" + fmt.Sprintf("%v", params[k])
	}
	if api != "" {
		plainText = "/" + api + "&" + plainParms[1:]
	} else {
		plainText = plainParms[1:]
	}

	plainText = url.QueryEscape(plainText)

	return plainText
}

func sign(plainText string, secretKey string) (sign string) {
	hmacObj := hmac.New(sha1.New, []byte(secretKey))
	hmacObj.Write([]byte(plainText))
	sign = base64.StdEncoding.EncodeToString(hmacObj.Sum(nil))
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
	var fi os.FileInfo
	fi, err = fin.Stat()
	if err != nil {
		return "", err
	}
	filesize := fi.Size()

	params := map[string]interface{}{}
	fileName := key

	//////
	api := "api/cos_upload"
	params["accessId"] = QcloudAccessID
	params["bucketId"] = QcloudBucket
	params["secretId"] = QcloudAccessKeyID
	params["cosFile"] = fileName
	params["path"] = "/"
	time := fmt.Sprintf("%v", time.Now().Unix())
	params["time"] = time
	/////

	fmt.Println("params[\"accessId\"]:", QcloudAccessID)
	fmt.Println("params[\"bucketId\"]:", QcloudBucket)
	fmt.Println("params[\"secretId\"]:", QcloudAccessKeyID)
	fmt.Println("params[\"cosFile\"]:", fileName)

	uploadPlainText := makePlainText(api, params)

	sign := sign(uploadPlainText, QcloudAccessKeySecret)

	var requstUrl string
	requstUrl = "http://" + QcloudEndpoint + "/" + api + "?bucketId=" + QcloudBucket + "&cosFile=" + fileName + "&path=%2F" + "&accessId=" + QcloudAccessID + "&secretId=" + QcloudAccessKeyID + "&time=" + time + "&sign=" + sign

	req, _ := http.NewRequest("POST", requstUrl, fin)
	req.Body = fin
	req.ContentLength = filesize
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return "", err
	}

	downloadUrl := qcloudGetDownloadUrl(fileName)

	return downloadUrl, nil
}

func qcloudGetDownloadUrl(fileName string) (downloadUrl string) {

	params := map[string]interface{}{}

	//////
	params["accessId"] = QcloudAccessID
	params["bucketId"] = QcloudBucket
	params["secretId"] = QcloudAccessKeyID
	params["path"] = "/" + fileName
	time := fmt.Sprintf("%v", time.Now().Unix())
	params["time"] = time

	downloadPlainText := makePlainText("", params)
	sign := sign(downloadPlainText, QcloudAccessKeySecret)
	url := "cos.myqcloud.com/" + QcloudAccessID + "/" + QcloudBucket + "/" + fileName + "?" + "secretId=" + QcloudAccessKeyID + "&time=" + time
	url += "&sign=" + sign
	return url

}
