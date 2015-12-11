package analyzer

import "sync"

const (
	channelSize = 200
)

type ShareChannel struct {
	In         chan string
	OutSuccess chan string
	OutFailure chan string
	ExitFlag   bool
	waitGroup  *sync.WaitGroup
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
	/*go func() {
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
	}()*/
}

func (sc *ShareChannel) Close() {
	sc.ExitFlag = true
	sc.waitGroup.Wait()

	/*for f := true; f; {
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
	}*/

	close(sc.In)
	close(sc.OutSuccess)
	close(sc.OutFailure)
}
