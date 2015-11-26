package backend

import (
	"fmt"

	"github.com/containerops/dockyard/backend/drivers"
	_"github.com/containerops/dockyard/backend/drivers/aliyun"
	_"github.com/containerops/dockyard/backend/drivers/amazons3"
	_ "github.com/containerops/dockyard/backend/drivers/native"
	_ "github.com/containerops/dockyard/backend/drivers/qiniu"
	_ "github.com/containerops/dockyard/backend/drivers/upyun"
	"github.com/containerops/wrench/setting"
)

var Sc drivers.ShareChannel

func InitBackend() error {
	if initfunc, existed := drivers.Drv[setting.BackendDriver]; existed {
		initfunc()
	} else {
		return fmt.Errorf("Driver %v is not registered", setting.BackendDriver)
	}

	//Init goroutine
	Sc = *drivers.NewShareChannel()
	Sc.Open()

	return nil
}
