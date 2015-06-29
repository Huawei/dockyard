package backend

import (
	"errors"
	"reflect"
)

type Injector map[string]reflect.Value

func NewInjector(size int) Injector {
	return make(Injector, size)

}

func (inj Injector) Bind(name string, fn interface{}) {
	v := reflect.ValueOf(fn)
	inj[name] = v
}

func (inj Injector) Call(name string, params ...interface{}) (result []reflect.Value, err error) {
	if _, ok := inj[name]; !ok {
		err = errors.New(name + " does not exist.")
		return
	}
	if len(params) != inj[name].Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = inj[name].Call(in)
	return
}
