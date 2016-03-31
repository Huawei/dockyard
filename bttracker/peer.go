package bttracker

import "time"

// Peer represents a bittorrent peer
type Peer struct {
	ID         string `json:"id,omitempty" bencode:"id,omitempty"`
	Port       int    `json:"port,omitempty" bencode:"port,omitempty"`
	Uploaded   int64
	Downloaded int64
	Left       int64
	Event      string //started,stoped,completed
	Compact    int
	NoPeerId   int    //Indicates that the tracker can omit peer id field in peers dictionary. This option is ignored if compact is enabled.
	Numwant    int    //Optional,This value is permitted to be zero. If omitted, typically defaults to 50 peers.
	IP         string //Optional
	Key        string //Optional
	TrackerId  string //Optional
	Expires    time.Time
}
