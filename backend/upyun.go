package backend

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/upyun/go-sdk/upyun"
)

var (
	g_upEndpoint string
	g_upBucket   string
	g_upUser     string
	g_upPasswd   string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	err := upgetconfig(gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Errorf("read config file conf/runtime.conf fail:" + err.Error())
		os.Exit(1)
	}

	g_injector.Bind("upcloudsave", upcloudsave)
}

func upgetconfig(conffile string) (err error) {
	var conf config.ConfigContainer
	conf, err = config.NewConfig("ini", conffile)
	if err != nil {
		return err
	}

	g_upEndpoint = conf.String("upCloud::endpoint")
	if g_upEndpoint == "" {
		return errors.New("read config file's endpoint failed!")
	}

	g_upBucket = conf.String("upCloud::bucket")
	if g_upBucket == "" {
		return errors.New("read config file's bucket failed!")
	}

	g_upUser = conf.String("upCloud::user")
	if g_upUser == "" {
		return errors.New("read config file's user failed!")
	}

	g_upPasswd = conf.String("upCloud::passwd")
	if g_upPasswd == "" {
		return errors.New("read config file's passwd failed!")

	}
	return nil
}

func upcloudsave(file string) (url string, err error) {

	var key string
	//get the filename from the file , eg,get "1.txt" from "/home/liugenping/1.txt"
	for _, key = range strings.Split(file, "/") {

	}
	opath := "/" + g_upBucket + "/" + key
	url = "http://" + g_upEndpoint + opath

	var u *upyun.UpYun
	u = upyun.NewUpYun(g_upBucket, g_upUser, g_upPasswd)
	if nil == u {
		return "", errors.New("UpYun.NewUpYun Fail")
	}

	/*	Endpoint list:
		Auto    = "v0.api.upyun.com"
		Telecom = "v1.api.upyun.com"
		Cnc     = "v2.api.upyun.com"
		Ctt     = "v3.api.upyun.com"
	*/
	u.SetEndpoint(g_upEndpoint)

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
