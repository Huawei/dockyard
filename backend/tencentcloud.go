package backend

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
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
	g_tencentEndpoint        string
	g_tencentAccessID        string
	g_tencentBucket          string
	g_tencentAccessKeyID     string
	g_tencentAccessKeySecret string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	err := tencentgetconfig(gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Errorf("read config file conf/runtime.conf fail:" + err.Error())
		os.Exit(0)
	}

	// 用户定义变量
	/*
		g_tencentEndpoint = "cosapi.myqcloud.com"

		g_tencentAccessID = "11000464"
		g_tencentBucket = "test"
		g_tencentAccessKeySecret = "4ceCa4wNP10c40QPPDgXdfx5MhvuCBWG"
		g_tencentAccessKeyID = "AKIDBxM1SkbDzdEtLED1KeQhW8HjW5qRu2R5"
	*/

	g_injector.Bind("tencentcloudsave", tencentcloudsave)

}

func tencentgetconfig(conffile string) (err error) {
	var conf config.ConfigContainer
	conf, err = config.NewConfig("ini", conffile)
	if err != nil {
		return err
	}

	g_tencentEndpoint = conf.String("tencentcloud::endpoint")
	if g_tencentEndpoint == "" {
		return errors.New("read config file's endpoint failed!")
	}

	g_tencentAccessID = conf.String("tencentcloud::accessID")
	if g_tencentAccessID == "" {
		return errors.New("read config file's accessID failed!")
	}

	g_tencentBucket = conf.String("tencentcloud::bucket")
	if g_tencentBucket == "" {
		return errors.New("read config file's bucket failed!")
	}

	g_tencentAccessKeyID = conf.String("tencentcloud::accessKeyID")
	if g_tencentAccessKeyID == "" {
		return errors.New("read config file's accessKeyID failed!")
	}

	g_tencentAccessKeySecret = conf.String("tencentcloud::accessKeysecret")
	if g_tencentAccessKeySecret == "" {
		return errors.New("read config file's accessKeysecret failed!")
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

func tencentcloudsave(file string) (url string, err error) {

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
	params["accessId"] = g_tencentAccessID
	params["bucketId"] = g_tencentBucket
	params["secretId"] = g_tencentAccessKeyID
	params["cosFile"] = fileName
	params["path"] = "/"
	time := fmt.Sprintf("%v", time.Now().Unix())
	params["time"] = time
	/////

	uploadPlainText := makePlainText(api, params)

	sign := sign(uploadPlainText, g_tencentAccessKeySecret)

	var requstUrl string
	requstUrl = "http://" + g_tencentEndpoint + "/" + api + "?bucketId=" + g_tencentBucket + "&cosFile=" + fileName + "&path=%2F" + "&accessId=" + g_tencentAccessID + "&secretId=" + g_tencentAccessKeyID + "&time=" + time + "&sign=" + sign

	req, _ := http.NewRequest("POST", requstUrl, fin)
	req.Body = fin
	req.ContentLength = filesize
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return "", err
	}

	downloadUrl := tencentGetDownloadUrl(fileName)

	return downloadUrl, nil
}

func tencentGetDownloadUrl(fileName string) (downloadUrl string) {

	params := map[string]interface{}{}

	//////
	params["accessId"] = g_tencentAccessID
	params["bucketId"] = g_tencentBucket
	params["secretId"] = g_tencentAccessKeyID
	params["path"] = "/" + fileName
	time := fmt.Sprintf("%v", time.Now().Unix())
	params["time"] = time

	downloadPlainText := makePlainText("", params)
	sign := sign(downloadPlainText, g_tencentAccessKeySecret)
	url := "cos.myqcloud.com/" + g_tencentAccessID + "/" + g_tencentBucket + "/" + fileName + "?" + "secretId=" + g_tencentAccessKeyID + "&time=" + time
	url += "&sign=" + sign
	return url

}
