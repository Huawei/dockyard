package clair

import (
	"sync"

	"github.com/coreos/clair/database"
)

const (
	channelSize = 200
)

type Input struct {
	Manifest []byte
	Vendor   string
	Version  string
}

type ShareChannel struct {
	In         chan Input
	OutSuccess chan []*database.Vulnerability
	OutFailure chan error
	ExitFlag   bool
	waitGroup  *sync.WaitGroup
}

func NewShareChannel() *ShareChannel {
	return &ShareChannel{
		make(chan Input, channelSize),
		make(chan []*database.Vulnerability, channelSize),
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

func (sc *ShareChannel) putOutSuccess(vulns []*database.Vulnerability) {
	sc.OutSuccess <- vulns
}

func (sc *ShareChannel) GutOutSuccess() []*database.Vulnerability {
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
			if err := Put(in.Manifest, in.Vendor, in.Version); err != nil {
				sc.putOutFailure(err)
			} else {
				if vulns, err := Get(in.Manifest, in.Vendor, in.Version); err != nil {
					sc.putOutFailure(err)
				} else {
					sc.putOutSuccess(vulns)
				}
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
			if err := Put(obj.Manifest, obj.Vendor, obj.Version); err != nil {
				sc.putOutFailure(err)
			} else {
				if vulns, err := Get(obj.Manifest, obj.Vendor, obj.Version); err != nil {
					sc.putOutFailure(err)
				} else {
					sc.putOutSuccess(vulns)
				}
			}
		default:
			f = false
		}
	}

	close(sc.In)
	close(sc.OutSuccess)
	close(sc.OutFailure)
}
