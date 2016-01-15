package amazons3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

func init() {
	drivers.Register("amazons3", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("amazons3save", amazons3save)
}

func amazons3save(file string) (url string, err error) {

	var key string

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

	requstUrl := "http://" + setting.Bucket + "." + setting.Endpoint + "/" + key
	r, _ := http.NewRequest("PUT", requstUrl, fin)
	r.ContentLength = int64(filesize)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("X-Amz-Acl", "public-write")

	fmt.Println(requstUrl)

	amazons3Sign(r, key, setting.AccessKeyID, setting.AccessKeysecret)
	_, err = http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}

	url = "http://" + setting.Endpoint + "/" + setting.Bucket + "/" + key
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

	plainText += "/" + setting.Bucket + "/" + key
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
