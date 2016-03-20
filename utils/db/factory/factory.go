package factory

import (
	"fmt"
)

type DRFactory interface {
	RegisterModel(models ...interface{})
	InitDB(driver, user, passwd, uri, name string, partition int64) error
	Insert(obj interface{}) error
	Update(obj interface{}, params ...string) error
	Get(obj interface{}, params ...string) (bool, error)
	Save(obj interface{}, params ...string) error
}

var Drivers = make(map[string]DRFactory)

func Register(name string, instance DRFactory) error {
	if _, existed := Drivers[name]; existed {
		return fmt.Errorf("%v has already been registered", name)
	}

	Drivers[name] = instance
	return nil
}
