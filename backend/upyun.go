package backend

import (
	"errors"
	"os"
	"strings"

	"github.com/upyun/go-sdk/upyun"
)

func upyunsave(file string) (url string, err error) {

	var tempkey string
	for _, tempkey = range strings.Split(file, "/") {

	}
	opath := "/" + BUCKETNAME + "/" + tempkey
	tempUrl := "http://" + ENDPOINT + opath

	var u *upyun.UpYun
	u = upyun.NewUpYun(BUCKETNAME, USER, PASSWD)
	if nil == u {
		return "", errors.New("UpYun.NewUpYun Fail")
	}

	fin, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fin.Close()

	var str string

	str, err = u.Put(tempkey, fin, false, "")
	if err != nil {
		return "", err
	}
	return tempUrl, nil

}
