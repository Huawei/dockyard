package backend

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/astaxie/beego/config"
)

var (
	DRIVER          string
	ENDPOINT        string
	BUCKETNAME      string
	AccessKeyID     string
	AccessKeySecret string
	//password for upyun
	USER   string
	PASSWD string
)

type InputObject struct {
	Key        string `json:"key"`
	Uploadfile string `json:"uploadfile"`
}

type OutputObject struct {
	Key         string `json:"key"`
	Uploadfile  string `json:"uploadfile"`
	Downloadurl string `json:"downloadurl"`
}

/**
input:
	 file, eg."/home/liugenping/gowsp/src/github.com/containerops/dockyard/2.png"
return:
	key, eg."2.png"
*/

func init() {

	err := getconfile("conf/runtime.conf")
	if err != nil {
		fmt.Errorf("读取配置文件 conf/runtime.conf 错误: %v", err)
	}
}

func getconfile(file string) (err error) {
	var tmperr error
	var conf config.ConfigContainer

	conf, tmperr = config.NewConfig("ini", file)
	if tmperr != nil {
		return tmperr
	}

	DRIVER = conf.String("backenddriver")
	if DRIVER == "" {
		return errors.New("read config file's backenddriver failed!")
	}

	ENDPOINT = conf.String(DRIVER + "::endpoint")
	if ENDPOINT == "" {
		return errors.New("read config file's endpoint failed!")
	}

	BUCKETNAME = conf.String(DRIVER + "::bucket")
	if BUCKETNAME == "" {
		return errors.New("read config file's bucket failed!")
	}

	if DRIVER == "up" {
		USER = conf.String(DRIVER + "::usr")
		if USER == "" {
			return errors.New("read config file's usr failed!")
		}
		PASSWD = conf.String(DRIVER + "::passwd")
		if PASSWD == "" {
			return errors.New("read config file's passwd failed!")
		}

	} else {

		AccessKeyID = conf.String(DRIVER + "::accessKeyID")
		if AccessKeyID == "" {
			return errors.New("read config file's accessKeyID failed!")
		}

		AccessKeySecret = conf.String(DRIVER + "::accessKeysecret")
		if AccessKeySecret == "" {
			return errors.New("read config file's accessKeysecret failed!")
		}

	}
	return nil
}

func Save(inputJson string) (outJson string, err error) {

	var tmpErr error
	var url string
	inputObj := InputObject{}

	tmpErr = json.Unmarshal([]byte(inputJson), &inputObj)
	if nil != tmpErr {
		return "", tmpErr
	}

	switch DRIVER {
	case "qiniu":
		url, tmpErr = qiniusave(inputObj.Uploadfile)
	case "up":
		url, tmpErr = qiniusave(inputObj.Uploadfile)
	case "ali":
		url, tmpErr = alisave(inputObj.Uploadfile)
	default:
		return "", errors.New("no saving place is config")
	}

	if nil != tmpErr {
		return "", tmpErr

	}

	outputObj := &OutputObject{Key: inputObj.Key, Uploadfile: inputObj.Uploadfile, Downloadurl: url}
	tempOutJson, tmpErr := json.Marshal(outputObj)
	if err != nil {
		return "", tmpErr
	}
	return string(tempOutJson), nil
}
