package upyun

import (
	"errors"
	"os"
	"strings"

	"github.com/upyun/go-sdk/upyun"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

func init() {
	drivers.Register("upyun", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("upyunsave", upyunsave)
}

func upyunsave(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}

	opath := "/" + setting.Bucket + "/" + key
	url = "http://" + setting.Endpoint + opath

	var u *upyun.UpYun
	u = upyun.NewUpYun(setting.Bucket, setting.User, setting.Passwd)
	if nil == u {
		return "", errors.New("UpYun.NewUpYun Fail")
	}

	u.SetEndpoint(setting.Endpoint)

	fin, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fin.Close()

	_, err = u.Put(key, fin, false, "")
	if err != nil {
		return "", err
	}
	return url, nil
}
