package backend

import (
	"strings"

	"github.com/qiniu/api/conf"
	"github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
)

func qiniusave(file string) (url string, err error) {

	conf.ACCESS_KEY = AccessKeyID
	conf.SECRET_KEY = AccessKeySecret
	var tempkey string
	for _, tempkey = range strings.Split(file, "/") {

	}

	tempUrl := "http://" + ENDPOINT + "/" + tempkey

	putPolicy := rs.PutPolicy{Scope: BUCKETNAME}
	uptoken := putPolicy.Token(nil)

	var ret io.PutRet
	var extra = &io.PutExtra{}
	err = io.PutFile(nil, &ret, uptoken, tempkey, file, extra)
	if err != nil {
		return "", err
	} else {
		return tempUrl, nil
	}

}
