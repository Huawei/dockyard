package analyzer

import (
	"encoding/json"
	"sync"
)

const (
	channelSize = 128
)

const (
	LOCAL = iota
	REGISTRY
	MANIFEST
)

type InLocal struct {
	FileName string
	FilePath string
}

type InRegistry struct {
	ImgURL   string
	Username string
	Passwd   string
	Insecure bool
}

type InManifest struct {
	JsonIn string
}

type Input struct {
	typeIn int
	jsonIn string
}

type ShareChannel struct {
	In         chan Input
	OutSuccess chan string
	OutFailure chan error
	ExitFlag   bool
	waitGroup  *sync.WaitGroup
}

func NewShareChannel() *ShareChannel {
	return &ShareChannel{
		make(chan Input, channelSize),
		make(chan string, channelSize),
		make(chan error, channelSize),
		false,
		new(sync.WaitGroup),
	}
}

func (sc *ShareChannel) PutIn(in Input) {
	sc.In <- in
}

func (sc *ShareChannel) getIn() Input {
	return <-sc.In
}

func (sc *ShareChannel) putOutSuccess(attrs string) {
	sc.OutSuccess <- attrs
}

func (sc *ShareChannel) GetOutSuccess() string {
	return <-sc.OutSuccess
}

func (sc *ShareChannel) putOutFailure(err error) {
	sc.OutFailure <- err
}

func (sc *ShareChannel) GutOutFailure() error {
	return <-sc.OutFailure
}

func (sc *ShareChannel) Open() {
	sc.waitGroup.Add(1)
	go func() {
		for !sc.ExitFlag {
			in := sc.getIn()
			jsonOut, err := analyseRun(in.typeIn, in.jsonIn)
			if nil != err {
				sc.putOutFailure(err)
			} else {
				sc.putOutSuccess(jsonOut)
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
			jsonOut, err := analyseRun(obj.typeIn, obj.jsonIn)
			if nil != err {
				sc.putOutFailure(err)
			} else {
				sc.putOutSuccess(jsonOut)
			}
		default:
			f = false
		}
	}

	close(sc.In)
	close(sc.OutSuccess)
	close(sc.OutFailure)
}

func analyseRun(typeIn int, jsonIn string) (string, error) {
	switch typeIn {
	case LOCAL:
		in := InLocal{}
		err := json.Unmarshal([]byte(jsonIn), &in)
		if nil != err {
			return "", err
		}
		imgAttr, err := AnalyseLocal(in.FileName, in.FilePath)
		if nil != err {
			return "", err
		} else {
			jsonTmp, err := json.Marshal(imgAttr)
			if nil != err {
				return "", err
			} else {
				return string(jsonTmp), nil
			}
		}
	case REGISTRY:
		in := InRegistry{}
		err := json.Unmarshal([]byte(jsonIn), &in)
		if nil != err {
			return "", err
		}
		imgAttr, err := AnalyseRegistry(in.ImgURL, in.Username, in.Passwd, in.Insecure)
		if nil != err {
			return "", err
		} else {
			jsonTmp, err := json.Marshal(imgAttr)
			if nil != err {
				return "", err
			} else {
				return string(jsonTmp), nil
			}
		}
	case MANIFEST:
		in := InManifest{}
		err := json.Unmarshal([]byte(jsonIn), &in)
		if nil != err {
			return "", err
		}
		imgAttr, err := AnalyseManifestFile(in.JsonIn)
		if nil != err {
			return "", err
		} else {
			jsonTmp, err := json.Marshal(imgAttr)
			if nil != err {
				return "", err
			} else {
				return string(jsonTmp), nil
			}
		}
	}

	return "", nil
}
