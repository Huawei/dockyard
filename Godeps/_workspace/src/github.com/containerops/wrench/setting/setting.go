package setting

import (
	"fmt"

	"github.com/astaxie/beego/config"
)

const (
	APIVERSION_V1 = iota
	APIVERSION_V2
	APIVERSION_ACI
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
	OssSwitch           string
)

// object storage driver config parameters
// TBD: It should be considered to refine the universal config parameters
var (
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeysecret string

	//upyun unique
	User   string
	Passwd string

	//qcloud unique
	AccessID string

	//googlecloud unique
	Projectid      string
	Scope          string
	PrivateKeyFile string
	Clientemail    string
)

// Clair service config parameters
var (
	//Path of the database. Default:  '/db'
	ClairDBPath string
	//Remove all the data in DB after stop the clair service. Default: false
	ClairKeepDB bool
	//Log level of the clair lib. Default: 'info'
	//All values: ['critical, error, warning, notice, info, debug, trace']
	ClairLogLevel string
	//Update CVE date in every '%dh%dm%ds'. Default: '1h0m0s'
	ClairUpdateDuration string
	//Return CVEs with minimal priority to Dockyard. Default: 'Low'
	//All values: ['Unknown, Negligible, Low, Medium, High, Critical, Defcon1']
	ClairVulnPriority string
)

// OSS backend driver parameters
var (
	APIPort      int
	APIHttpsPort int
	PartSizeMB   int
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
	if ossswitch := conf.String("dockyard::ossswitch"); ossswitch != "" {
		OssSwitch = ossswitch
	} else if ossswitch == "" {
		OssSwitch = "disable"
	}

	//Dockyard object storage,default to use dockyard storage
	BackendDriver = "native"
	if backenddriver := conf.String("dockyard::driver"); backenddriver != "" {
		BackendDriver = backenddriver
	}

	// TBD: It should be considered to refine the universal config parameters
	switch BackendDriver {
	case "native":
		//It will be supported soon
	case "qiniu", "aliyun", "amazons3":
		if endpoint := conf.String(BackendDriver + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(BackendDriver + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if accessKeyID := conf.String(BackendDriver + "::" + "accessKeyID"); accessKeyID != "" {
			AccessKeyID = accessKeyID
		} else {
			err = fmt.Errorf("AccessKeyID value is null")
		}

		if accessKeysecret := conf.String(BackendDriver + "::" + "accessKeysecret"); accessKeysecret != "" {
			AccessKeysecret = accessKeysecret
		} else {
			err = fmt.Errorf("AccessKeysecret value is null")
		}

	case "upyun":
		if endpoint := conf.String(BackendDriver + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(BackendDriver + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if user := conf.String(BackendDriver + "::" + "user"); user != "" {
			User = user
		} else {
			err = fmt.Errorf("User value is null")
		}

		if passwd := conf.String(BackendDriver + "::" + "passwd"); passwd != "" {
			Passwd = passwd
		} else {
			err = fmt.Errorf("Passwd value is null")
		}

	case "qcloud":
	//It will be supported soon
	case "oss":
		APIPort, err = conf.Int(BackendDriver + "::" + "apiport")
		APIHttpsPort, err = conf.Int(BackendDriver + "::" + "apihttpsport")
		PartSizeMB, err = conf.Int(BackendDriver + "::" + "partsizemb")
	case "googlecloud":
		//It will be supported soon
	default:
		err = fmt.Errorf("Doesn't support %v now", BackendDriver)
	}

	ClairDBPath = conf.String("clair::path")
	ClairLogLevel = conf.String("clair::logLevel")
	ClairKeepDB, _ = conf.Bool("clair::keepDB")
	ClairUpdateDuration = conf.String("clair::updateDuration")
	ClairVulnPriority = conf.String("clair::vulnPriority")

	return err
}
