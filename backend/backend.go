package backend

import (
	"github.com/containerops/dockyard/backend/drivers"
	_ "github.com/containerops/dockyard/backend/drivers/aliyun"
	_ "github.com/containerops/dockyard/backend/drivers/amazons3"
	_ "github.com/containerops/dockyard/backend/drivers/googlecloud"
	_ "github.com/containerops/dockyard/backend/drivers/oss"
	_ "github.com/containerops/dockyard/backend/drivers/qcloud"
	_ "github.com/containerops/dockyard/backend/drivers/qiniu"
	_ "github.com/containerops/dockyard/backend/drivers/upyun"
	"github.com/containerops/wrench/setting"
)

var Sc drivers.ShareChannel

func InitBackend() {
	initfunc, existed := drivers.Drv[setting.BackendDriver]
	if !existed {
		return
	}
	initfunc()

	//Init goroutine
	Sc = *drivers.NewShareChannel()
	Sc.Open()
}
