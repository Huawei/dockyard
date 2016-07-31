/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package setting

import (
	"fmt"
	"os"

	"github.com/astaxie/beego/config"
	"github.com/ngaut/log"
)

var (
	//@Global Config

	//AppName should be "dockyard"
	AppName string
	//Usage is short description
	Usage   string
	Version string
	Author  string
	Email   string

	//@Basic Runtime Config

	RunMode        string
	ListenMode     string
	HttpsCertFile  string
	HttpsKeyFile   string
	LogPath        string
	LogLevel       string
	DatabaseDriver string
	DatabaseURI    string
	Domains        string

	//@Docker V1 Config

	DockerStandalone      string
	DockerRegistryVersion string

	//@Docker V2 Config

	DockerDistributionVersion string
)

//
func init() {
	if err := setConfig("conf/containerops.conf"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//set log level
	log.SetLevelByString(LogLevel)
}

func setConfig(path string) error {
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		return fmt.Errorf("Read %s error: %v", path, err.Error())
	}

	//config globals
	if appname := conf.String("appname"); appname != "" {
		AppName = appname
	} else if appname == "" {
		return fmt.Errorf("AppName config value is null")
	}

	if usage := conf.String("usage"); usage != "" {
		Usage = usage
	} else if usage == "" {
		return fmt.Errorf("Usage config value is null")
	}

	if version := conf.String("version"); version != "" {
		Version = version
	} else if version == "" {
		return fmt.Errorf("Version config value is null")
	}

	if author := conf.String("author"); author != "" {
		Author = author
	} else if author == "" {
		return fmt.Errorf("Author config value is null")
	}

	if email := conf.String("email"); email != "" {
		Email = email
	} else if email == "" {
		return fmt.Errorf("Email config value is null")
	}

	//config runtime
	if runmode := conf.String("runmode"); runmode != "" {
		RunMode = runmode
	} else if runmode == "" {
		return fmt.Errorf("RunMode config value is null")
	}

	if listenmode := conf.String("listenmode"); listenmode != "" {
		ListenMode = listenmode
	} else if listenmode == "" {
		return fmt.Errorf("ListenMode config value is null")
	}

	if httpscertfile := conf.String("httpscertfile"); httpscertfile != "" {
		HttpsCertFile = httpscertfile
	} else if httpscertfile == "" {
		return fmt.Errorf("HttpsCertFile config value is null")
	}

	if httpskeyfile := conf.String("httpskeyfile"); httpskeyfile != "" {
		HttpsKeyFile = httpskeyfile
	} else if httpskeyfile == "" {
		return fmt.Errorf("HttpsKeyFile config value is null")
	}

	if logpath := conf.String("log::filepath"); logpath != "" {
		LogPath = logpath
	} else if logpath == "" {
		return fmt.Errorf("LogPath config value is null")
	}

	if loglevel := conf.String("log::level"); loglevel != "" {
		LogLevel = loglevel
	} else if loglevel == "" {
		return fmt.Errorf("LogLevel config value is null")
	}

	if databasedriver := conf.String("database::driver"); databasedriver != "" {
		DatabaseDriver = databasedriver
	} else if databasedriver == "" {
		return fmt.Errorf("Database Driver config value is null")
	}

	if databaseuri := conf.String("database::uri"); databaseuri != "" {
		DatabaseURI = databaseuri
	} else if databaseuri == "" {
		return fmt.Errorf("Database URI config vaule is null")
	}

	if domains := conf.String("deployment::domain"); domains != "" {
		Domains = domains
	} else if domains == "" {
		return fmt.Errorf("Deployment domains value is null")
	}

	//TODO: Add a config option for provide Docker Registry V1.
	//TODO: Link @middle/header/setRespHeaders, @handler/dockerv1/-functions.
	if standalone := conf.String("dockerv1::standalone"); standalone != "" {
		DockerStandalone = standalone
	} else if standalone == "" {
		return fmt.Errorf("DockerV1 standalone value is null")
	}

	if registry := conf.String("dockerv1::version"); registry != "" {
		DockerRegistryVersion = registry
	} else if registry == "" {
		return fmt.Errorf("DockerV1 Registry Version value is null")
	}

	if distribution := conf.String("dockerv2::distribution"); distribution != "" {
		DockerDistributionVersion = distribution
	} else if distribution == "" {
		return fmt.Errorf("DockerV2 Distribution Version value is null")
	}

	return nil
}
