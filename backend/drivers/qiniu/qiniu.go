package qiniu

import (
	"strings"

	"github.com/qiniu/api.v6/conf"
	"github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

func init() {
	drivers.Register("qiniu", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("qiniusave", qiniusave)
}

func qiniusave(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}

	conf.ACCESS_KEY = setting.AccessKeyID
	conf.SECRET_KEY = setting.AccessKeysecret

	url = "http://" + setting.Endpoint + "/" + key

	putPolicy := rs.PutPolicy{Scope: setting.Bucket}
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
