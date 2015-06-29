package backend

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/qiniu/api/conf"
	"github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
)

var (
	g_qiniuEndpoint        string
	g_qiniuBucket          string
	g_qiniuAccessKeyID     string
	g_qiniuAccessKeySecret string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	err := qiniugetconfig(gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Errorf("read config file conf/runtime.conf fail:" + err.Error())
		os.Exit(0)
	}

	conf.ACCESS_KEY = g_qiniuAccessKeyID
	conf.SECRET_KEY = g_qiniuAccessKeySecret

	g_injector.Bind("qiniucloudsave", qiniucloudsave)
}

func qiniugetconfig(conffile string) (err error) {
	var conf config.ConfigContainer
	conf, err = config.NewConfig("ini", conffile)
	if err != nil {
		return err
	}

	g_qiniuEndpoint = conf.String("qiniucloud::endpoint")
	if g_qiniuEndpoint == "" {
		return errors.New("read config file's endpoint failed!")
	}

	g_qiniuBucket = conf.String("qiniucloud::bucket")
	if g_qiniuBucket == "" {
		return errors.New("read config file's bucket failed!")
	}

	g_qiniuAccessKeyID = conf.String("qiniucloud::accessKeyID")
	if g_qiniuAccessKeyID == "" {
		return errors.New("read config file's accessKeyID failed!")
	}

	g_qiniuAccessKeySecret = conf.String("qiniucloud::accessKeysecret")
	if g_qiniuAccessKeySecret == "" {
		return errors.New("read config file's accessKeysecret failed!")
	}
	return nil
}

func qiniucloudsave(file string) (url string, err error) {

	var key string
	//get the filename from the file , eg,get "1.txt" from /home/liugenping/1.txt
	for _, key = range strings.Split(file, "/") {

	}

	url = "http://" + g_qiniuEndpoint + "/" + key

	putPolicy := rs.PutPolicy{Scope: g_qiniuBucket}
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
