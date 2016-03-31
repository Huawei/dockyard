package handler

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/bttracker"
)

func getAnnounceInput(ctx *macaron.Context) (string, *bttracker.Peer, error) {
	info_hash := ctx.Query("info_hash")
	if len(info_hash) <= 0 {
		return "", nil, fmt.Errorf("input info_hash error")
	}

	peer_id := ctx.Query("peer_id")
	if len(peer_id) <= 0 {
		return "", nil, fmt.Errorf("input peer_id error")
	}
	port := ctx.QueryInt("port")
	if port <= 0 {
		return "", nil, fmt.Errorf("input port error")
	}
	uploaded := ctx.QueryInt64("uploaded")
	if uploaded < 0 {
		return "", nil, fmt.Errorf("input uploaded error")
	}
	downloaded := ctx.QueryInt64("downloaded")
	if downloaded < 0 {
		return "", nil, fmt.Errorf("input downloaded error")
	}
	left := ctx.QueryInt64("left")
	if left < 0 {
		return "", nil, fmt.Errorf("input left error")
	}
	compact := ctx.QueryInt("compact")
	if compact < 0 || compact > 1 {
		return "", nil, fmt.Errorf("input compact error")
	}
	event := ctx.Query("event")
	if !(strings.EqualFold(event, "started") ||
		strings.EqualFold(event, "stoped") ||
		strings.EqualFold(event, "completed")) {
		return "", nil, fmt.Errorf("input event error")
	}

	no_peer_id := ctx.QueryInt("no_peer_id")

	ip := ctx.Query("ip")
	if len(ip) <= 0 {
		ip = ctx.RemoteAddr()
	}
	numwant := ctx.QueryInt("numwant")
	if numwant <= 0 {
		numwant = 50
	}
	key := ctx.Query("key")
	trackerid := ctx.Params(":trackerid")

	p := &bttracker.Peer{
		ID:         peer_id,
		Port:       port,
		Uploaded:   uploaded,
		Downloaded: downloaded,
		Left:       left,
		Event:      event,
		Compact:    compact,
		NoPeerId:   no_peer_id,
		Numwant:    numwant,
		IP:         ip,
		Key:        key,
		TrackerId:  trackerid,
	}
	return info_hash, p, nil
}

// Handle client's request for tracker
func AnnounceHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	const bencodErrFormat = "d8:failure reason%d:%se"
	infoHash, peer, err := getAnnounceInput(ctx)
	if err != nil {
		return 200, []byte(fmt.Sprintf(bencodErrFormat, len(err.Error()), err.Error()))
	}

	if resp, err := bttracker.TS.Announce(infoHash, peer); err != nil {
		return 200, []byte(fmt.Sprintf(bencodErrFormat, len(err.Error()), err.Error()))
	} else {
		return 200, resp
	}
}
