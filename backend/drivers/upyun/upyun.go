package upyun

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/upyun/go-sdk/upyun"

	"github.com/containerops/dockyard/backend/drivers"
)

var (
	UpyunEndpoint string
	UpyunBucket   string
	UpyunUser     string
	UpyunPasswd   string
)

type UpyunDrv struct{}

func init() {
	drivers.Drv["upyun"] = &UpyunDrv{}
}

func (d *UpyunDrv) ReadConfig(conf config.ConfigContainer) error {
	UpyunEndpoint = conf.String("upyun::endpoint")
	if UpyunEndpoint == "" {
		return fmt.Errorf("Read endpoint of Upyun failed!")
	}

	UpyunBucket = conf.String("upyun::bucket")
	if UpyunBucket == "" {
		return fmt.Errorf("Read bucket of Upyun failed!")
	}

	UpyunUser = conf.String("upyun::user")
	if UpyunUser == "" {
		return fmt.Errorf("Read user of Upyun failed!")
	}

	UpyunPasswd = conf.String("upyun::passwd")
	if UpyunPasswd == "" {
		return fmt.Errorf("Read passwd of Upyun failed!")
	}

	drivers.InjectReflect.Bind("upyunsave", upyunsave)
	return nil
}

func upyunsave(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}

	opath := "/" + UpyunBucket + "/" + key
	url = "http://" + UpyunEndpoint + opath

	var u *upyun.UpYun
	u = upyun.NewUpYun(UpyunBucket, UpyunUser, UpyunPasswd)
	if nil == u {
		return "", errors.New("UpYun.NewUpYun Fail")
	}

	u.SetEndpoint(UpyunEndpoint)

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
