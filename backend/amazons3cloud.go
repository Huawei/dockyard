package backend

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego/config"
)

var (
	g_amazons3Endpoint        string
	g_amazons3Bucket          string
	g_amazons3AccessKeyID     string
	g_amazons3AccessKeySecret string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	err := amazons3getconfig(gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Errorf("read config file conf/runtime.conf fail:" + err.Error())
		os.Exit(1)
	}
	g_injector.Bind("amazons3cloudsave", amazons3cloudsave)
}

func amazons3getconfig(conffile string) (err error) {
	conf, err := config.NewConfig("ini", conffile)
	if err != nil {
		return err
	}

	g_amazons3Endpoint = conf.String("amazons3cloud::endpoint")
	if g_amazons3Endpoint == "" {
		return errors.New("read config file's endpoint failed!")
	}

	g_amazons3Bucket = conf.String("amazons3cloud::bucket")
	if g_amazons3Bucket == "" {
		return errors.New("read config file's bucket failed!")
	}

	g_amazons3AccessKeyID = conf.String("amazons3cloud::accessKeyID")
	if g_amazons3AccessKeyID == "" {
		return errors.New("read config file's accessKeyID failed!")
	}

	g_amazons3AccessKeySecret = conf.String("amazons3cloud::accessKeysecret")
	if g_amazons3AccessKeySecret == "" {
		return errors.New("read config file's accessKeysecret failed!")
	}
	return nil
}

func amazons3cloudsave(file string) (url string, err error) {

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

	requstUrl := "http://" + g_amazons3Bucket + "." + g_amazons3Endpoint + "/" + key
	r, _ := http.NewRequest("PUT", requstUrl, fin)
	r.ContentLength = int64(filesize)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("X-Amz-Acl", "public-read")

	amazons3Sign(r, key, g_amazons3AccessKeyID, g_amazons3AccessKeySecret)
	_, err = http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	url = "http://" + g_amazons3Endpoint + "/" + g_amazons3Bucket + "/" + key
	return url, nil

}

func amazons3Sign(r *http.Request, key string, accessKeyId string, accessKeySecret string) {

	plainText := amazons3cloudMakePlainText(r, key)
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	h.Write([]byte(plainText))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	r.Header.Set("Authorization", "AWS "+accessKeyId+":"+sign)
}

func amazons3cloudMakePlainText(r *http.Request, key string) (plainText string) {

	plainText = r.Method + "\n"
	plainText += r.Header.Get("content-md5") + "\n"
	plainText += r.Header.Get("content-type") + "\n"
	if _, ok := r.Header["X-Amz-Date"]; !ok {
		plainText += r.Header.Get("date") + "\n"
	}

	amzHeader := getAmzHeaders(r)
	if amzHeader != "" {
		plainText += amzHeader + "\n"
	}

	plainText += "/" + g_amazons3Bucket + "/" + key
	return
}

func getAmzHeaders(r *http.Request) (amzHeader string) {
	var keys []string
	for k, _ := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-amz-") {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)
	var a []string
	for _, k := range keys {
		v := r.Header[k]
		a = append(a, strings.ToLower(k)+":"+strings.Join(v, ","))
	}
	for _, h := range a {

		return h
	}
	return ""
}
