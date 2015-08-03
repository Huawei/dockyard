package setting

import (
	"fmt"

	"github.com/astaxie/beego/config"
)

var (
	conf          config.ConfigContainer
	AppName       string
	Usage         string
	Version       string
	Author        string
	Email         string
	RunMode       string
	ListenMode    string
	HttpsCertFile string
	HttpsKeyFile  string
	LogPath       string
	DBURI         string
	DBPasswd      string
	DBDB          int64
	BasePath      string
)

const (
	APIVERSION_V1 = iota
	APIVERSION_V2
)

func init() {
	var err error

	conf, err = config.NewConfig("ini", "conf/containerops.conf")
	if err != nil {
		fmt.Errorf("read conf/dockyard.conf failure: %v", err)
	}

	if appname := conf.String("appname"); appname != "" {
		AppName = appname
	}

	if usage := conf.String("usage"); usage != "" {
		Usage = usage
	}

	if version := conf.String("version"); version != "" {
		Version = version
	}

	if author := conf.String("author"); author != "" {
		Author = author
	}

	if email := conf.String("email"); email != "" {
		Email = email
	}

	if runmode := conf.String("runmode"); runmode != "" {
		RunMode = runmode
	}

	if listenmode := conf.String("listenmode"); listenmode != "" {
		ListenMode = listenmode
	}

	if httpscertfile := conf.String("httpscertfile"); httpscertfile != "" {
		HttpsCertFile = httpscertfile
	}

	if httpskeyfile := conf.String("httpskeyfile"); httpskeyfile != "" {
		HttpsKeyFile = httpskeyfile
	}

	if logpath := conf.String("log::filepath"); logpath != "" {
		LogPath = logpath
	}

	if dburi := conf.String("db::uri"); dburi != "" {
		DBURI = dburi
	}

	DBPasswd = conf.String("db::passwd")
	DBDB, _ = conf.Int64("db::db")

	if DockerBasePath := conf.String("docker::BasePath"); DockerBasePath != "" {
		BasePath = DockerBasePath
	}
}
