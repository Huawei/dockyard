package db

import (
	"fmt"

	"github.com/containerops/wrench/db/factory"
	_ "github.com/containerops/wrench/db/orm"
	_ "github.com/containerops/wrench/db/redis"
)

var Drv factory.DRFactory

func RegisterDriver(name string) error {
	if Drv != nil {
		return fmt.Errorf("Only support one driver at one time")
	}

	for k, v := range factory.Drivers {
		if k == name && v != nil {
			Drv = factory.Drivers[k]
			return nil
		}
	}

	return fmt.Errorf("Not support driver %v", name)
}
