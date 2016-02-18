package web

import (
	"fmt"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/backend"
	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/oss"
	"github.com/containerops/dockyard/router"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

func SetDockyardMacaron(m *macaron.Macaron) {
	if err := db.RegisterDriver(setting.DBDriver); err != nil {
		fmt.Printf("Register database driver error: %s\n", err.Error())
	} else {
		db.Drv.RegisterModel(new(models.Repository), new(models.Tag), new(models.Image))
		err := db.Drv.InitDB(setting.DBDriver, setting.DBUser, setting.DBPasswd, setting.DBURI, setting.DBName, setting.DBDB)
		if err != nil {
			fmt.Printf("Connect database error: %s\n", err.Error())
		}
	}

	backend.InitBackend()

	if err := middleware.Initfunc(); err != nil {
		fmt.Printf("Init middleware error: %s\n", err.Error())
	}

	//Setting Middleware
	middleware.SetMiddlewares(m)

	//Setting Router
	router.SetRouters(m)

	//Start Object Storage Service if sets in conf
	if strings.EqualFold(setting.OssSwitch, "enable") {
		ossobj := oss.Instance()
		ossobj.StartOSS()
	}

}
