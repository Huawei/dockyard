package web

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/backend"
	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/oss"
	"github.com/containerops/dockyard/router"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

func SetDockyardMacaron(m *macaron.Macaron) {
	//Setting Database
	if err := db.InitDB(setting.DBURI, setting.DBPasswd, setting.DBDB); err != nil {
		fmt.Printf("Connect Database error %s", err.Error())
	}

	if err := backend.InitBackend(); err != nil {
		fmt.Printf("Init backend error %s", err.Error())
	}

	if err := middleware.Initfunc(); err != nil {
		fmt.Printf("Init middleware error %s", err.Error())
	}

	//Setting Middleware
	middleware.SetMiddlewares(m)

	//Setting Router
	router.SetRouters(m)

	//Start Object Storage Service if sets in conf
	if strings.EqualFold(setting.BackendDriver, "oss") {
		ossobj := oss.Instance()
		ossobj.StartOSS()
	}

	//Create acpool to store aci/asc/pubkey
	err := func() error {
		acpoolname := setting.ImagePath + "/acpool"
		if _, err := os.Stat(acpoolname); err == nil {
			return nil
		}

		if err := os.Mkdir(acpoolname, 0755); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		fmt.Printf("Create acpool for rkt failed %s", err.Error())
	}
}
