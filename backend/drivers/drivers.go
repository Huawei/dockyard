package drivers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/containerops/wrench/setting"
)

var channelSize = 200
var InjectReflect = NewInjector(50)

type In struct {
	Key        string `json:"key"`
	Uploadfile string `json:"uploadfile"`
}

type OutSuccess struct {
	Key         string `json:"key"`
	Uploadfile  string `json:"uploadfile"`
	Downloadurl string `json:"downloadurl"`
}

type ShareChannel struct {
	In         chan string
	OutSuccess chan string
	OutFailure chan string
	ExitFlag   bool
	waitGroup  *sync.WaitGroup
}

type INITFUNC func()

var Drv = make(map[string]INITFUNC)

func Register(name string, initfunc INITFUNC) error {
	if _, existed := Drv[name]; existed {
		return fmt.Errorf("%v has already been registered", name)
	}

	Drv[name] = initfunc

	return nil
}

func NewShareChannel() *ShareChannel {
	return &ShareChannel{
		make(chan string, channelSize),
		make(chan string, channelSize),
		make(chan string, channelSize),
		false,
		new(sync.WaitGroup),
	}
}

func (sc *ShareChannel) PutIn(jsonObj string) {
	sc.In <- jsonObj
}

func (sc *ShareChannel) getIn() (jsonObj string) {
	return <-sc.In
}

func (sc *ShareChannel) putOutSuccess(jsonObj string) {
	sc.OutSuccess <- jsonObj
}

func (sc *ShareChannel) GutOutSuccess() (jsonObj string) {
	return <-sc.OutSuccess
}

func (sc *ShareChannel) putOutFailure(jsonObj string) {
	sc.OutFailure <- jsonObj
}

func (sc *ShareChannel) GutOutFailure() (jsonObj string) {
	return <-sc.OutFailure
}

func (sc *ShareChannel) Open() {
	sc.waitGroup.Add(1)
	go func() {
		for !sc.ExitFlag {
			obj := sc.getIn()
			outJson, err := Save(obj)
			if nil != err {
				sc.putOutFailure(obj)
			} else {
				sc.putOutSuccess(outJson)
			}
		}
		sc.waitGroup.Done()
	}()
}

func (sc *ShareChannel) Close() {
	sc.ExitFlag = true
	sc.waitGroup.Wait()

	for f := true; f; {
		select {
		case obj := <-sc.In:
			outJson, err := Save(obj)
			if nil != err {
				sc.putOutFailure(obj)
			} else {
				sc.putOutSuccess(outJson)
			}
		default:
			f = false
		}
	}

	close(sc.In)
	close(sc.OutSuccess)
	close(sc.OutFailure)
}

func Save(jsonIn string) (jsonOut string, err error) {

	var url string
	var rt []reflect.Value
	in := In{}
	var jsonTempOut []byte

	err = json.Unmarshal([]byte(jsonIn), &in)
	if nil != err {
		return "", err
	}

	rt, err = InjectReflect.Call(setting.BackendDriver+"save", in.Uploadfile)
	if nil != err {
		return "", err
	}

	if len(rt) > 1 && !rt[1].IsNil() {
		errstr := rt[1].MethodByName("Error").Call(nil)[0].String()
		if errstr != "" {
			return "", fmt.Errorf(errstr)
		}

	}
	url = rt[0].String()

	outSuccess := &OutSuccess{Key: in.Key, Uploadfile: in.Uploadfile, Downloadurl: url}
	jsonTempOut, err = json.Marshal(outSuccess)
	if err != nil {
		return "", err
	}

	return string(jsonTempOut), nil
}
