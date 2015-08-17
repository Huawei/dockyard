package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/astaxie/beego/config"
)

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

var channelSize = 200

//reflect struct
var g_injector = NewInjector(50)
var g_driver string

func init() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Println("read env GOPATH fail")
		os.Exit(1)
	}
	conf, err := config.NewConfig("ini", gopath+"/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Println(fmt.Errorf("read conf/runtime.conf fail: %v", err).Error())
		os.Exit(1)
	}

	g_driver = conf.String("backend::driver")
	if g_driver == "" {
		fmt.Println("read config file's dirver failed!")
		os.Exit(1)
	}
}

func NewShareChannel() *ShareChannel {
	return &ShareChannel{make(chan string, channelSize),
		make(chan string, channelSize),
		make(chan string, channelSize), false, new(sync.WaitGroup)}
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
				//fmt.Println(err)
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
				//fmt.Println(err)
				sc.putOutFailure(obj)
			} else {
				sc.putOutSuccess(outJson)
			}
		default:
			f = false
		}
	}

	close(sc.In)
	//close(sc.OutSuccess)
	//close(sc.OutFail)
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

	rt, err = g_injector.Call(g_driver+"save", in.Uploadfile)
	if nil != err {
		return "", err
	}

	if !rt[1].IsNil() {
		errstr := rt[1].MethodByName("Error").Call(nil)[0].String()
		if errstr != "" {
			return "", errors.New(errstr)
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
