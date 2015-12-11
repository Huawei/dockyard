package analyzer

import (
	"sync"

	"github.com/containerops/dockyard/analyzer/attr"
)

const (
	channelSize = 200
)

type Input struct {
	dockerURL string
	filePath  string
}

type ShareChannel struct {
	In         chan Input
	OutSuccess chan []attr.DockerImg_Attr
	OutFailure chan error
	ExitFlag   bool
	waitGroup  *sync.WaitGroup
}

func NewShareChannel() *ShareChannel {
	return &ShareChannel{
		make(chan Input, channelSize),
		make(chan []attr.DockerImg_Attr, channelSize),
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

func (sc *ShareChannel) putOutSuccess(attrs []attr.DockerImg_Attr) {
	sc.OutSuccess <- attrs
}

func (sc *ShareChannel) GutOutSuccess() []attr.DockerImg_Attr {
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
			attrs, err := AnalyseLocal(in.dockerURL, in.filePath)
			if nil != err {
				sc.putOutFailure(err)
			} else {
				sc.putOutSuccess(attrs)
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
			attrs, err := AnalyseLocal(obj.dockerURL, obj.filePath)
			if nil != err {
				sc.putOutFailure(err)
			} else {
				sc.putOutSuccess(attrs)
			}
		default:
			f = false
		}
	}

	close(sc.In)
	close(sc.OutSuccess)
	close(sc.OutFailure)
}
