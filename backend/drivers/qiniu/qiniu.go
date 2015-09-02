package qiniu

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/qiniu/api.v6/conf"
	"github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"

	"github.com/containerops/dockyard/backend/drivers"
)

var (
	QiniuEndpoint        string
	QiniuBucket          string
	QiniuAccessKeyID     string
	QiniuAccessKeySecret string
)

type QiniuDrv struct{}

func init() {
	drivers.Drv["qiniu"] = &QiniuDrv{}
}

func (d *QiniuDrv) ReadConfig(conf config.ConfigContainer) error {

	QiniuEndpoint = conf.String("qiniu::endpoint")
	if QiniuEndpoint == "" {
		return fmt.Errorf("Read endpoint of Qiniu failed!")
	}

	QiniuBucket = conf.String("qiniu::bucket")
	if QiniuBucket == "" {
		return fmt.Errorf("Read bucket of Qiniu failed!")
	}

	QiniuAccessKeyID = conf.String("qiniu::accessKeyID")
	if QiniuAccessKeyID == "" {
		return fmt.Errorf("Read accessKeyID of Qiniu failed!")
	}

	QiniuAccessKeySecret = conf.String("qiniu::accessKeysecret")
	if QiniuAccessKeySecret == "" {
		return fmt.Errorf("Read accessKeysecret of Qiniu failed!")
	}

	drivers.InjectReflect.Bind("qiniusave", qiniusave)
	return nil
}

func qiniusave(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}

	conf.ACCESS_KEY = QiniuAccessKeyID
	conf.SECRET_KEY = QiniuAccessKeySecret

	url = "http://" + QiniuEndpoint + "/" + key

	putPolicy := rs.PutPolicy{Scope: QiniuBucket}
	uptoken := putPolicy.Token(nil)

	var ret io.PutRet
	var extra = &io.PutExtra{}
	err = io.PutFile(nil, &ret, uptoken, key, file, extra)
	if err != nil {
		return "", err
	} else {
		return url, nil
	}

}
