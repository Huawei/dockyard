package backend

import (
	"strings"

	"github.com/yanunon/oss-go-api/oss"
)

func alisave(file string) (url string, err error) {

	client := oss.NewClient(ENDPOINT, AccessKeyID, AccessKeySecret, 0)

	var tempkey string
	for _, tempkey = range strings.Split(file, "/") {

	}
	opath := "/" + BUCKETNAME + "/" + tempkey
	tempUrl := "http://" + ENDPOINT + opath

	temperr := client.PutObject(opath, file)
	if nil != temperr {
		return "", temperr
	} else {
		return tempUrl, nil
	}
}
