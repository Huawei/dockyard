package qiniu

import (
	"github.com/qiniu/api.v6/conf"
	"github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"

	"github.com/containerops/dockyard/backend/factory"
	"github.com/containerops/dockyard/utils/setting"
)

type qiniudesc struct{}

func init() {
	factory.Register("qiniu", &qiniudesc{})
}

func (q *qiniudesc) Save(file string) (url string, err error) {
	conf.ACCESS_KEY = setting.AccessKeyID
	conf.SECRET_KEY = setting.AccessKeysecret

	url = "http://" + setting.Endpoint + "/" + file

	putPolicy := rs.PutPolicy{Scope: setting.Bucket}
	uptoken := putPolicy.Token(nil)

	var ret io.PutRet
	var extra = &io.PutExtra{}
	err = io.PutFile(nil, &ret, uptoken, file, file, extra)
	if err != nil {
		return "", err
	} else {
		return url, nil
	}

}

func (q *qiniudesc) Get(file string) ([]byte, error) {
	return []byte(""), nil
}

func (q *qiniudesc) Delete(file string) error {
	return nil
}
