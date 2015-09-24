package backend

import (
	"github.com/containerops/dockyard/backend/drivers"
	_ "github.com/containerops/dockyard/backend/drivers/native"
	_ "github.com/containerops/dockyard/backend/drivers/qiniu"
	_ "github.com/containerops/dockyard/backend/drivers/upyun"
)

var Sc drivers.ShareChannel

func InitBackend() {
	//Init goroutine
	Sc = *drivers.NewShareChannel()
	Sc.Open()
}
