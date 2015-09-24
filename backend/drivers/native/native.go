package native

import (
	"github.com/astaxie/beego/config"

	"github.com/containerops/dockyard/backend/drivers"
)

type NativeDrv struct{}

func init() {
	drivers.Drv["native"] = &NativeDrv{}
}

func (d *NativeDrv) ReadConfig(conf config.ConfigContainer) error {
	return nil
}
