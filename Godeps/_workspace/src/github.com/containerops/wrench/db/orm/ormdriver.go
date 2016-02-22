package orm

import (
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
	"github.com/huawei-openlab/newdb/orm"

	"github.com/containerops/wrench/db/factory"
)

type ormdrv struct{}

func init() {
	factory.Register("mysql", &ormdrv{})
}

func (od *ormdrv) RegisterModel(models ...interface{}) {
	orm.RegisterModel(models...)
}

func (od *ormdrv) InitDB(driver, user, passwd, uri, name string, partition int64) error {
	ds := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", user, passwd, uri, name)
	if err := orm.RegisterDataBase("default", driver, ds, 0, 0); err != nil {
		return err
	}

	if err := orm.RunSyncdb("default", false, false); err != nil {
		return err
	}

	return nil
}

func (od *ormdrv) Get(obj interface{}, params ...string) (bool, error) {
	n := len(params)
	if n <= 0 {
		return false, fmt.Errorf("Invalid key")
	}

	keys := []string{}
	s := reflect.ValueOf(obj).Elem()
	typeOfS := s.Type()
	for i := 0; i < n; i++ {
		for k := 0; k < s.NumField(); k++ {
			f := s.Field(k)

			if f.Interface() == params[i] {
				keys = append(keys, typeOfS.Field(k).Name)
				break
			}
		}
	}

	if len(keys) <= 0 {
		return false, fmt.Errorf("Wrong key to query")
	}

	o := orm.NewOrm()
	if err := o.Read(obj, keys...); err != nil {
		if err == orm.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func (od *ormdrv) Save(obj interface{}, params ...string) error {
	o := orm.NewOrm()

	exists, err := od.Get(obj, params...)
	if err != nil {
		return err
	}

	if err := o.Begin(); err != nil {
		return err
	}

	if !exists {
		_, err = o.Insert(obj)
	} else {
		_, err = o.Update(obj)
	}

	if err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}

	return err
}

func (od *ormdrv) Insert(obj interface{}) error {
	o := orm.NewOrm()

	err := o.Begin()
	if err != nil {
		return err
	}

	if _, err := o.Insert(obj); err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}

	return err
}

func (od *ormdrv) Update(obj interface{}, params ...string) error {
	o := orm.NewOrm()

	err := o.Begin()
	if err != nil {
		return err
	}

	if _, err := o.Update(obj, params...); err != nil {
		o.Rollback()
	} else {
		o.Commit()
	}

	return err
}
