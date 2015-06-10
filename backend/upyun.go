package backend

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/upyun/go-sdk/upyun"
)

/**
input:
	 file, eg."/home/liugenping/gowsp/src/github.com/containerops/dockyard/2.png"
return:
	key, eg."2.png"
*/

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
	fmt.Println("u.Put=" + str)
	return tempUrl, nil

}
