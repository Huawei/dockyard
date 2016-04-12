package bttracker

import (
	"testing"
	"time"
)

func Test_TrackerServer_Announce(t *testing.T) {

	TS.interval = 20
	infoHash1 := "hash1111111"
	infoHash2 := "hash2222222"
	p1 := &Peer{
		ID:         "111111",
		Port:       2000,
		Uploaded:   30,
		Downloaded: 500,
		Left:       600,
		Event:      "started", //started,stoped,completed
		Compact:    0,
		NoPeerId:   0,
		Numwant:    0,
		IP:         "192.168.0.2", //Optional
		Key:        "",            //Optional
		TrackerId:  "",            //Optional
		Expires:    time.Now().Add(time.Duration(TS.interval) * time.Second * time.Minute),
	}

	p2 := &Peer{
		ID:         "222222",
		Port:       2000,
		Uploaded:   30,
		Downloaded: 500,
		Left:       600,
		Event:      "started", //started,stoped,completed
		Compact:    0,
		NoPeerId:   0,
		Numwant:    0,
		IP:         "192.168.0.2", //Optional
		Key:        "",            //Optional
		TrackerId:  "",            //Optional
		Expires:    time.Now().Add(time.Duration(TS.interval) * time.Second * time.Minute),
	}
	TS.addPeer(infoHash1, p1)
	TS.addPeer(infoHash1, p2)
	TS.addPeer(infoHash2, p1)
	TS.addPeer(infoHash2, p2)

	p3 := &Peer{
		ID:         "333333",
		Port:       2000,
		Uploaded:   30,
		Downloaded: 500,
		Left:       600,
		Event:      "started", //started,stoped,completed
		Compact:    0,
		NoPeerId:   0,
		Numwant:    0,
		IP:         "192.168.0.2", //Optional
		Key:        "",            //Optional
		TrackerId:  "",            //Optional
		Expires:    time.Now().Add(time.Duration(TS.interval) * time.Second * time.Minute),
	}
	if resp, err := TS.Announce(infoHash1, p3); err != nil {
		t.Error(err)
	} else {
		t.Log(string(resp))
	}
}
