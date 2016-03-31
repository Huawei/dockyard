package bttracker

import (
	"time"

	"github.com/liugenping/bencode"
)

var TS *TrackerServer

type TrackerServer struct {
	id string
	//Minimum announce interval.
	//If present clients must not reannounce more frequently than this.
	//MinInterval int
	//Interval in seconds that the client should wait between sending regular requests to the tracker
	interval int
	//info_hash to []*Peer
	torrent map[string]*PeerRepository
}

func init() {
	TS = NewTrackerServer()
}

func NewTrackerServer() *TrackerServer {
	return &TrackerServer{
		id:       "DOCKYARD TRACKER",
		interval: 60, //1 minutes
		torrent:  map[string]*PeerRepository{},
	}
}

func (s *TrackerServer) addPeer(infoHash string, p *Peer) {
	peerRepo, exist := s.torrent[infoHash]
	if exist {
		peerRepo.AddPeer(p)
	} else {
		pp := NewPeerRepository()
		pp.AddPeer(p)
		s.torrent[infoHash] = pp
	}
}

func (s *TrackerServer) deletePeer(infoHash string, p *Peer) {
	peerRepo, exist := s.torrent[infoHash]
	if exist {
		peerRepo.DeletePeer(p)
	}
}

func (s *TrackerServer) getPeers(infoHash string) []*Peer {
	peerRepo, exist := s.torrent[infoHash]
	if exist {
		return peerRepo.GetPeers()
	}
	return []*Peer{}
}

func (s *TrackerServer) getComplete(infoHash string) int {
	peerRepo, exist := s.torrent[infoHash]
	if exist {
		return peerRepo.complete
	} else {
		return 0
	}
}

func (s *TrackerServer) getIncomplete(infoHash string) int {
	peerRepo, exist := s.torrent[infoHash]
	if exist {
		return peerRepo.incomplete
	} else {
		return 0
	}
}

func (s *TrackerServer) getResp(infoHash string) ([]byte, error) {
	//failure reason:
	//tracker id:
	//complete:
	//incomplete:
	//peers:
	//////peer id:
	//////ip:
	//////port:
	resp := map[string]interface{}{}
	respPeers := []map[string]interface{}{}
	resp["tracker id"] = s.id
	resp["complete"] = s.getComplete(infoHash)
	resp["incomplete"] = s.getIncomplete(infoHash)
	peers := s.getPeers(infoHash)
	if len(peers) > 0 {
		for _, p := range peers {
			respPeer := map[string]interface{}{}
			respPeer["peer_id"] = p.ID
			respPeer["ip"] = p.IP
			respPeer["port"] = p.Port
			respPeers = append(respPeers, respPeer)
		}
		resp["peers"] = respPeers
	}
	if bytes, err := bencode.Marshal(resp); err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}

func (s *TrackerServer) Announce(infoHash string, peer *Peer) ([]byte, error) {
	if peer.Event == "stoped" {
		s.deletePeer(infoHash, peer)
		return []byte{}, nil
	}

	if resp, err := s.getResp(infoHash); err != nil {
		return nil, err
	} else {
		peer.Expires = time.Now().Add(time.Duration(s.interval) * time.Second)
		s.addPeer(infoHash, peer)
		return resp, nil
	}
}
