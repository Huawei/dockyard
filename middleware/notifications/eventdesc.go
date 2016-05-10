package notifications

import (
	"fmt"
	"time"
)

const (
	ManifestMediaType = "application/vnd.docker.distribution.manifest.v1+json"
	BlobMediaType     = "application/vnd.docker.distribution.manifest.v1+json" // TODO: to be confirm with receiver
	DefaultMedisType  = "application/octet-stream"
	EventsMediaType   = "application/vnd.docker.distribution.events.v1+json"
)

const (
	EventActionPull = "pull"
	EventActionPush = "push"
)

type Envelope struct {
	Events []Event `json:"events,omitempty"`
}

type Descriptor struct {
	MediaType string `json:"mediaType,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Digest    string `json:"digest,omitempty"`
}

type TargetCtx struct {
	Descriptor
	Length     int64     `json:"length,omitempty"`
	Repository string    `json:"repository,omitempty"`
	Tag        string    `json:"tag,omitempty"`
	URL        string    `json:"url,omitempty"`
	FSLayers   []FSLayer `json:"fsLayers,omitempty"`
}

type Event struct {
	ID        string        `json:"id,omitempty"`
	Timestamp time.Time     `json:"timestamp,omitempty"`
	Action    string        `json:"action,omitempty"`
	Target    TargetCtx     `json:"target,omitempty"`
	Req       RequestRecord `json:"request,omitempty"`
	Actor     ActorRecord   `json:"actor,omitempty"`
	Source    SourceRecord  `json:"source,omitempty"`
}

type ActorRecord struct {
	Name string `json:"name,omitempty"`
}

type RequestRecord struct {
	ID        string `json:"id"`
	Addr      string `json:"addr,omitempty"`
	Host      string `json:"host,omitempty"`
	Method    string `json:"method"`
	UserAgent string `json:"useragent"`
}

type SourceRecord struct {
	Addr       string `json:"addr,omitempty"`
	InstanceID string `json:"instanceID,omitempty"`
}

var (
	ErrSinkClosed = fmt.Errorf("sink: closed")
)

type Error struct {
	Err        error
	StatusCode int
}

type Sink interface {
	Write(events ...Event) Error
	Close() error
}

type Manifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	Name          string    `json:"name"`
	Tag           string    `json:"tag"`
	Architecture  string    `json:"architecture"`
	FSLayers      []FSLayer `json:"fsLayers"`
	History       []History `json:"history"`
}

type SignedManifest struct {
	Manifest
	Raw []byte `json:"-"`
}

type FSLayer struct {
	BlobSum string `json:"blobSum"`
}

type History struct {
	V1Compatibility string `json:"v1Compatibility"`
}
