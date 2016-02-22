package qiniu

import (
	"os"
	"strings"

	"golang.org/x/net/context"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"

	"github.com/containerops/dockyard/backend/driver"
	"github.com/containerops/wrench/setting"
)

func init() {
	driver.Register("qiniu", InitFunc)
}

func InitFunc() {
	driver.InjectReflect.Bind("qiniusave", qiniusave)
}

func qiniusave(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}
	fil, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fil.Close()
	fi, err := fil.Stat()
	if err != nil {
		return "", err
	}
	fsize := fi.Size()

	kodo.SetMac(setting.AccessKeyID, setting.AccessKeysecret)

	zone := 0
	c := kodo.New(zone, nil) //Create a Client object

	bucket := setting.Bucket
	policy := &kodo.PutPolicy{
		Scope:   bucket, // restriction
		Expires: 3600,   // expired time of uptoken
	}
	uptoken := c.MakeUptoken(policy)

	domain := setting.Endpoint
	url = kodocli.MakeBaseUrl(domain, key)

	//file less than 128m
	if fsize < (2 << 27) {
		zone = 0
		uploader := kodocli.NewUploader(zone, nil)
		ctx := context.Background()
		err = uploader.PutFile(ctx, nil, uptoken, key, file, nil)
	} else {
		//file more than 128m
		zone = 0
		uploader := kodocli.NewUploader(zone, nil)
		ctx := context.Background()
		err = uploader.RputFile(ctx, nil, uptoken, key, file, nil)
	}

	if err != nil {
		return "", err
	} else {
		return url, nil
	}

}
