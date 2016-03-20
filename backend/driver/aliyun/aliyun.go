package aliyun

import (
	"os"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/containerops/dockyard/backend/driver"
	"github.com/containerops/dockyard/utils/setting"
)

func init() {
	driver.Register("aliyun", InitFunc)
}

func InitFunc() {
	driver.InjectReflect.Bind("aliyunsave", aliyunsave)
}

func aliyunsave(file string) (url string, err error) {

	client, err := oss.New(setting.Endpoint, setting.AccessKeyID, setting.AccessKeysecret)
	if err != nil {
		return "", err
	}

	bucket, err := client.Bucket(setting.Bucket)
	if err != nil {
		return "", err
	}

	var key string
	//get the filename from the file , eg,get "1.txt" from /home/liugenping/1.txt
	for _, key = range strings.Split(file, "/") {

	}
	opath := "/" + setting.Bucket + "/" + key
	url = "http://" + setting.Endpoint + opath

	fd, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	fi, err := fd.Stat()
	if err != nil {
		return "", err
	}
	fsize := fi.Size()

	if fsize < (1 << 27) {
		err = bucket.PutObject(key, fd)
	} else {
		err = bucket.UploadFile(key, file, 2<<22)
	}

	if nil != err {
		return "", err
	} else {
		return url, nil
	}
}
