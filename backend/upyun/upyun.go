package upyun

import (
	"errors"
	"os"
	"strings"

	"github.com/upyun/go-sdk/upyun"

	"github.com/containerops/dockyard/backend/factory"
	"github.com/containerops/dockyard/utils/setting"
)

type upyundesc struct{}

func init() {
	factory.Register("upyun", &upyundesc{})
}

func (u *upyundesc) Save(file string) (url string, err error) {
	var key string
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

	if fsize < (2 << 27) {
		uf := upyun.NewUpYunForm(setting.Bucket, setting.Secret)
		if nil == uf {
			return "", errors.New("UpYun.NewUpYunForm Fail")
		}
		options := map[string]string{}

		_, err = uf.Put(file, key, 3600, options)

	} else {
		ump := upyun.NewUpYunMultiPart(setting.Bucket, setting.Secret, 1024000)
		if nil == ump {
			return "", errors.New("UpYun.NewUpYunForm Fail")
		}
		options := map[string]interface{}{}

		_, err = ump.Put(file, key, 3600, options)
	}
	if err != nil {
		return "", err
	}
	return url, nil
}

func (u *upyundesc) Get(file string) ([]byte, error) {
	return []byte(""), nil
}

func (u *upyundesc) Delete(file string) error {
	return nil
}
