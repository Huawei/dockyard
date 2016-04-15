package backend

import (
	"fmt"

	_ "github.com/containerops/dockyard/backend/aliyun"
	"github.com/containerops/dockyard/backend/factory"
	_ "github.com/containerops/dockyard/backend/gcs"
	_ "github.com/containerops/dockyard/backend/oss"
	_ "github.com/containerops/dockyard/backend/qcloud"
	_ "github.com/containerops/dockyard/backend/qiniu"
	_ "github.com/containerops/dockyard/backend/rados"
	_ "github.com/containerops/dockyard/backend/s3"
	_ "github.com/containerops/dockyard/backend/upyun"
)

var Drv factory.DrvInterface

func RegisterDriver(driver string) error {
	var err error
	if Drv != nil {
		return fmt.Errorf("Only support one driver at one time")
	}

	for k, v := range factory.Drivers {
		if k == driver && v != nil {
			Drv, err = factory.Drivers[k].New()
			if err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("Invalid driver %v", driver)
}
