package notifications

import (
	"bytes"
	"net/http"
	"time"

	"github.com/satori/go.uuid"

	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/signature"
)

type bridge struct {
	URL   string
	Actor ActorRecord
	Req   RequestRecord
	Sink  Sink
}

func newReqRecord(id string, r *http.Request) RequestRecord {
	return RequestRecord{
		ID:        id,
		Addr:      module.RemoteAddr(r),
		Host:      r.Host,
		Method:    r.Method,
		UserAgent: r.UserAgent(),
	}
}

func newBridge(url string, actor ActorRecord, req RequestRecord, sink Sink) *bridge {
	return &bridge{
		URL:   url,
		Actor: actor,
		Req:   req,
		Sink:  sink,
	}
}

func (b *bridge) createBlobEventAndWrite(action string, repo string, desc Descriptor) Error {
	event, err := b.createBlobEvent(action, repo, desc)
	if err != nil {
		return Error{err, http.StatusBadRequest}
	}

	return b.Sink.Write(*event)
}

func (b *bridge) createBlobEvent(action string, repo string, desc Descriptor) (*Event, error) {
	event := b.createEvent(action)
	event.Target.Descriptor = desc
	event.Target.Length = desc.Size
	event.Target.Repository = repo
	event.Target.Tag = ""

	if desc.Digest != "" {
		event.Target.URL = b.URL
	}

	return event, nil
}

func (b *bridge) createManifestEventAndWrite(action string, repo string, sm *SignedManifest) Error {
	manifestEvent, err := b.createManifestEvent(action, repo, sm)
	if err != nil {
		return Error{err, http.StatusBadRequest}
	}

	return b.Sink.Write(*manifestEvent)
}

func (b *bridge) createManifestEvent(action string, repo string, sm *SignedManifest) (*Event, error) {
	event := b.createEvent(action)
	event.Target.MediaType = ManifestMediaType
	event.Target.Repository = repo

	p, err := signature.Payload(sm.Raw)
	if err != nil {
		return nil, err
	}

	event.Target.Length = int64(len(p))
	event.Target.Digest, err = signature.FromReader(bytes.NewReader(p))
	if err != nil {
		return nil, err
	}

	event.Target.FSLayers = sm.FSLayers
	event.Target.URL = b.URL
	event.Target.Tag = sm.Tag

	return event, nil
}

func (b *bridge) createEvent(action string) *Event {
	event := &Event{
		ID:        utils.MD5(uuid.NewV4().String()),
		Timestamp: time.Now(),
		Action:    action,
	}
	//event.Source = b.source
	event.Actor = b.Actor
	event.Req = b.Req

	return event
}
