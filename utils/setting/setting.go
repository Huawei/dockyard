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
	conf config.Configer
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

	//DB
	DBDriver string
	DBUser   string
	DBPasswd string
	DBName   string
	DBURI    string
	DBDB     int64

	//Dockyard
	Backend             string
	ImagePath           string
	Domains             string
	RegistryVersion     string
	DistributionVersion string
	Standalone          string
	OssSwitch           string
)

// object storage driver config parameters
var (
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeysecret string

	//upyun unique
	Secret string

	//qcloud unique
	QcloudAccessID string

	//googlecloud unique
	Projectid          string
	Scope              string
	PrivateKeyFilePath string
	PrivateKeyFile     string
	Clientemail        string
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

	//config globals
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

	//config DB
	if dbdriver := conf.String("db::driver"); dbdriver != "" {
		DBDriver = dbdriver
	} else {
		err = fmt.Errorf("DB driver config value is null")
	}
	if dburi := conf.String("db::uri"); dburi != "" {
		DBURI = dburi
	}
	if dbuser := conf.String("db::user"); dbuser != "" {
		DBUser = dbuser
	}
	if dbpass := conf.String("db::passwd"); dbpass != "" {
		DBPasswd = dbpass
	}
	if dbname := conf.String("db::name"); dbname != "" {
		DBName = dbname
	}
	dbpartition, _ := conf.Int64("db::db")
	DBDB = dbpartition

	//config Dockyard
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
	if Backend = conf.String("dockyard::backend"); Backend != "" {
	}

	// TODO: It should be considered to refine the universal config parameters
	switch Backend {
	case "":
	case "qiniu", "aliyun", "s3":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if accessKeyID := conf.String(Backend + "::" + "accessKeyID"); accessKeyID != "" {
			AccessKeyID = accessKeyID
		} else {
			err = fmt.Errorf("AccessKeyID value is null")
		}

		if accessKeysecret := conf.String(Backend + "::" + "accessKeysecret"); accessKeysecret != "" {
			AccessKeysecret = accessKeysecret
		} else {
			err = fmt.Errorf("AccessKeysecret value is null")
		}
	case "upyun":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if secret := conf.String(Backend + "::" + "secret"); secret != "" {
			Secret = secret
		} else {
			err = fmt.Errorf("Secret value is null")
		}
	case "qcloud":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if accessID := conf.String(Backend + "::" + "accessID"); accessID != "" {
			QcloudAccessID = accessID
		} else {
			err = fmt.Errorf("accessID value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if accessKeyID := conf.String(Backend + "::" + "accessKeyID"); accessKeyID != "" {
			AccessKeyID = accessKeyID
		} else {
			err = fmt.Errorf("AccessKeyID value is null")
		}

		if accessKeysecret := conf.String(Backend + "::" + "accessKeysecret"); accessKeysecret != "" {
			AccessKeysecret = accessKeysecret
		} else {
			err = fmt.Errorf("AccessKeysecret value is null")
		}
	case "oss":
		APIPort, err = conf.Int(Backend + "::" + "apiport")
		APIHttpsPort, err = conf.Int(Backend + "::" + "apihttpsport")
		PartSizeMB, err = conf.Int(Backend + "::" + "partsizemb")
	case "gcs":
		if projectid := conf.String(Backend + "::" + "projectid"); projectid != "" {
			Projectid = projectid
		} else {
			err = fmt.Errorf("Projectid value is null")
		}

		if scope := conf.String(Backend + "::" + "scope"); scope != "" {
			Scope = scope
		} else {
			err = fmt.Errorf("Scope value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if keyfilepath := conf.String(Backend + "::" + "keyfilepath"); keyfilepath != "" {
			PrivateKeyFilePath = keyfilepath
		} else {
			err = fmt.Errorf("Privatekey value is null")
		}

		if privatekey := conf.String(Backend + "::" + "privatekey"); privatekey != "" {
			PrivateKeyFile = privatekey
		} else {
			err = fmt.Errorf("Privatekey value is null")
		}

		if clientemail := conf.String(Backend + "::" + "clientemail"); clientemail != "" {
			Clientemail = clientemail
		} else {
			err = fmt.Errorf("Clientemail value is null")
		}
	default:
		err = fmt.Errorf("Not support %v", Backend)
	}

	//Config of image security scanning
	ClairDBPath = conf.String("clair::path")
	ClairLogLevel = conf.String("clair::logLevel")
	ClairKeepDB, _ = conf.Bool("clair::keepDB")
	ClairUpdateDuration = conf.String("clair::updateDuration")
	ClairVulnPriority = conf.String("clair::vulnPriority")

	return err
}
