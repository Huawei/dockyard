package setting

import (
	"fmt"

	"github.com/astaxie/beego/config"
)

const (
	APIVERSION_V1 = iota
	APIVERSION_V2
)

var (
	conf config.ConfigContainer
)

var (
	//Global
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
	//Dockyard
	BackendDriver       string
	ImagePath           string
	Domains             string
	RegistryVersion     string
	DistributionVersion string
	Standalone          string
)

func SetConfig(path string) error {
	var err error

	conf, err = config.NewConfig("ini", path)
	if err != nil {
		return fmt.Errorf("Read %s error: %v", path, err.Error())
	}

	if appname := conf.String("appname"); appname != "" {
		AppName = appname
	} else if appname == "" {
		err = fmt.Errorf("AppName config value is null")
	}

	if usage := conf.String("usage"); usage != "" {
		Usage = usage
	} else if usage == "" {
		err = fmt.Errorf("Usage config value is null")
	}

	if version := conf.String("version"); version != "" {
		Version = version
	} else if version == "" {
		err = fmt.Errorf("Version config value is null")
	}

	if author := conf.String("author"); author != "" {
		Author = author
	} else if author == "" {
		err = fmt.Errorf("Author config value is null")
	}

	if email := conf.String("email"); email != "" {
		Email = email
	} else if email == "" {
		err = fmt.Errorf("Email config value is null")
	}

	if runmode := conf.String("runmode"); runmode != "" {
		RunMode = runmode
	} else if runmode == "" {
		err = fmt.Errorf("RunMode config value is null")
	}

	if listenmode := conf.String("listenmode"); listenmode != "" {
		ListenMode = listenmode
	} else if listenmode == "" {
		err = fmt.Errorf("ListenMode config value is null")
	}

	if httpscertfile := conf.String("httpscertfile"); httpscertfile != "" {
		HttpsCertFile = httpscertfile
	} else if httpscertfile == "" {
		err = fmt.Errorf("HttpsCertFile config value is null")
	}

	if httpskeyfile := conf.String("httpskeyfile"); httpskeyfile != "" {
		HttpsKeyFile = httpskeyfile
	} else if httpskeyfile == "" {
		err = fmt.Errorf("HttpsKeyFile config value is null")
	}

	if logpath := conf.String("log::filepath"); logpath != "" {
		LogPath = logpath
	} else if logpath == "" {
		err = fmt.Errorf("LogPath config value is null")
	}

	if dburi := conf.String("db::uri"); dburi != "" {
		DBURI = dburi
	} else if dburi == "" {
		err = fmt.Errorf("DBURI config value is null")
	}

	if dbpass := conf.String("db::passwd"); dbpass != "" {
		DBPasswd = dbpass
	}

	DBDB, err = conf.Int64("db::db")

	//Dockyard object storage,default to use dockyard storage
	BackendDriver = "native"
	if backenddriver := conf.String("dockyard::driver"); backenddriver != "" {
		BackendDriver = backenddriver
	}

	if imagepath := conf.String("dockyard::path"); imagepath != "" {
		ImagePath = imagepath
	} else if imagepath == "" {
		err = fmt.Errorf("Image path config value is null")
	}

	if domains := conf.String("dockyard::domains"); domains != "" {
		Domains = domains
	} else if domains == "" {
		err = fmt.Errorf("Domains value is null")
	}

	if registryVersion := conf.String("dockyard::registry"); registryVersion != "" {
		RegistryVersion = registryVersion
	} else if registryVersion == "" {
		err = fmt.Errorf("Registry version value is null")
	}

	if distributionVersion := conf.String("dockyard::distribution"); distributionVersion != "" {
		DistributionVersion = distributionVersion
	} else if distributionVersion == "" {
		err = fmt.Errorf("Distribution version value is null")
	}

	if standalone := conf.String("dockyard::standalone"); standalone != "" {
		Standalone = standalone
	} else if standalone == "" {
		err = fmt.Errorf("Standalone version value is null")
	}

	return err
}
