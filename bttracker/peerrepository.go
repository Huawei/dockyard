package bttracker

import (
	"sync"
	"time"
)

type PeerRepository struct {
	sync.Mutex
	complete   int
	incomplete int
	//peer's ID to Peer
	peers map[string]*Peer
}

func NewPeerRepository() *PeerRepository {
	return &PeerRepository{
		complete:   0,
		incomplete: 0,
		peers:      map[string]*Peer{}}
}

func (r *PeerRepository) AddPeer(p *Peer) {
	r.Lock()
	r.peers[p.ID] = p
	r.Unlock()
}

func (r *PeerRepository) DeletePeer(p *Peer) {
	r.Lock()
	delete(r.peers, p.ID)
	r.Unlock()
}

func (r *PeerRepository) GetPeers() []*Peer {
	peers := []*Peer{}
	r.Lock()
	for _, p := range r.peers {

		if p.Expires.After(time.Now()) { //peer is not expires
			if p.Left == 0 {
				r.complete = r.complete + 1
			} else {
				r.incomplete = r.incomplete + 1
			}
			peers = append(peers, p)
		} else { //peer is expires
			delete(r.peers, p.ID)
		}
	}
	r.Unlock()
	return peers
}
