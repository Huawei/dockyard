package backend

import (
	"github.com/containerops/dockyard/backend/driver"
	_ "github.com/containerops/dockyard/backend/driver/aliyun"
	_ "github.com/containerops/dockyard/backend/driver/amazons3"
	_ "github.com/containerops/dockyard/backend/driver/googlecloud"
	_ "github.com/containerops/dockyard/backend/driver/oss"
	_ "github.com/containerops/dockyard/backend/driver/qcloud"
	_ "github.com/containerops/dockyard/backend/driver/qiniu"
	_ "github.com/containerops/dockyard/backend/driver/upyun"
	"github.com/containerops/dockyard/utils/setting"
)

var Sc driver.ShareChannel

func InitBackend() {
	initfunc, existed := driver.Drv[setting.BackendDriver]
	if !existed {
		return
	}
	initfunc()

	//Init goroutine
	Sc = *driver.NewShareChannel()
	Sc.Open()
}
