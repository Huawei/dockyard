package backend

import (
	"strings"

	"github.com/qiniu/api/conf"
	"github.com/qiniu/api/io"
	"github.com/qiniu/api/rs"
)

var DOMAIN = "7xjdg0.com1.z0.glb.clouddn.com"

/**
input:
	 file, eg."/home/liugenping/gowsp/src/github.com/containerops/dockyard/2.png"
return:
	key, eg."2.png"
*/

func qiniusave(file string) (url string, err error) {

	//public domain
	conf.ACCESS_KEY = AccessKeyID
	conf.SECRET_KEY = AccessKeySecret
	var tempkey string
	for _, tempkey = range strings.Split(file, "/") {

	}

	tempUrl := "http://" + ENDPOINT + "/" + tempkey

	//1.genarate uptoken
	putPolicy := rs.PutPolicy{Scope: BUCKETNAME}
	uptoken := putPolicy.Token(nil)

	//2. upload
	var ret io.PutRet
	var extra = &io.PutExtra{}

	// ret       变量用于存取返回的信息，详情见 io.PutRet
	// uptoken   为业务服务器生成的上传口令
	// localFile 为本地文件名
	// extra     为上传文件的额外信息，详情见 io.PutExtra，可选
	err = io.PutFile(nil, &ret, uptoken, tempkey, file, extra)
	if err != nil {
		return "", err
	} else {
		return tempUrl, nil
	}

}
