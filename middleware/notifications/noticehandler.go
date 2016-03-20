package notifications

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/handler"
	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
	"github.com/containerops/dockyard/utils/signature"
)

type notification struct{}

var notice Sink

func init() {
	middleware.Register("notifications", middleware.HandlerInterface(&notification{}))
}

func (n *notification) InitFunc() error {
	var sinks []Sink

	for _, e := range setting.JSONConfCtx.Notifications.Endpoints {
		if e.Disabled {
			continue
		}

		endpoint := newEndpoint(EndpointDesc{
			Name:      e.Name,
			URL:       e.URL,
			Headers:   e.Headers,
			Timeout:   e.Timeout,
			Threshold: e.Threshold,
			Backoff:   e.Backoff,
			EventDB:   e.EventDB,
			Disabled:  e.Disabled,
		})
		sinks = append(sinks, endpoint)
	}

	notice = newSyncNotice(sinks...)

	return nil
}

func (n *notification) Handler(ctx *macaron.Context) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	//Notification function just supports DockerV2 now
	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil || exists == false {
		return
	}
	if r.Version != setting.APIVERSION_V2 {
		return
	}

	actor := ActorRecord{Name: namespace}
	repo := fmt.Sprintf("%v/%v", namespace, repository)

	switch ctx.Req.Method {
	case "HEAD":
		digest := ctx.Params(":digest")
		tarsum := strings.Split(digest, ":")[1]

		i := new(models.Image)
		if exists, _ := i.Get(tarsum); exists == false {
			return
		}

		desc := Descriptor{
			MediaType: DefaultMedisType,
			Size:      i.Size,
			Digest:    digest,
		}

		req := newReqRecord(utils.EncodeBasicAuth(namespace, "headblobsv2"), ctx.Req.Request)
		requrl, err := module.NewURLBuilderFromRequest(ctx.Req.Request).BuildBlobURL(repo, desc.Digest)
		if err != nil {
			fmt.Errorf("[REGISTRY API V2] Head blobs and get request URL failed, error:: %v", err.Error())
		}

		b := newBridge(requrl, actor, req, notice)
		Err := b.createBlobEventAndWrite(EventActionPull, repo, desc)
		if Err.Err != nil {
			ctx.Resp.WriteHeader(http.StatusForbidden)
			return
		} else if Err.StatusCode >= 300 {
			ctx.Resp.WriteHeader(Err.StatusCode)
			return
		}

	case "GET":
		if flag := utils.Compare(ctx.Req.RequestURI, "/v2/"); flag == 0 {
			return
		}

		if flag := strings.Contains(ctx.Req.RequestURI, "/blobs/"); flag == true {
			digest := ctx.Params(":digest")
			tarsum := strings.Split(digest, ":")[1]

			i := new(models.Image)
			if exists, _ := i.Get(tarsum); exists == false {
				return
			}

			desc := Descriptor{
				MediaType: BlobMediaType,
				Size:      i.Size,
				Digest:    digest,
			}

			req := newReqRecord(utils.EncodeBasicAuth(namespace, "getblobsv2"), ctx.Req.Request)
			requrl, err := module.NewURLBuilderFromRequest(ctx.Req.Request).BuildBlobURL(repo, desc.Digest)
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Get blobs and get request URL failed, error:: %v", err.Error())
			}

			b := newBridge(requrl, actor, req, notice)
			Err := b.createBlobEventAndWrite(EventActionPull, repo, desc)
			if Err.Err != nil {
				ctx.Resp.WriteHeader(http.StatusForbidden)
				return
			} else if Err.StatusCode >= 300 {
				ctx.Resp.WriteHeader(Err.StatusCode)
				return
			}
			return
		}

		if flag := strings.Contains(ctx.Req.RequestURI, "/manifests/"); flag == true {
			t := new(models.Tag)
			if exists, err := t.Get(namespace, repository, ctx.Params(":tag")); err != nil || !exists {
				return
			}

			digest, err := signature.DigestManifest([]byte(t.Manifest))
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())
				return
			}

			var sm SignedManifest
			if err := json.Unmarshal([]byte(t.Manifest), &sm); err != nil {
				fmt.Errorf("Unmarshal manifest error")
			}

			sm.Raw = make([]byte, len(t.Manifest), len(t.Manifest))
			copy(sm.Raw, []byte(t.Manifest))

			req := newReqRecord(utils.EncodeBasicAuth(namespace, "getmanifestv2"), ctx.Req.Request)
			requrl, err := module.NewURLBuilderFromRequest(ctx.Req.Request).BuildManifestURL(sm.Name, digest)
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Get manifest and get request URL failed, error:: %v", err.Error())
			}

			b := newBridge(requrl, actor, req, notice)
			Err := b.createManifestEventAndWrite(EventActionPull, repo, &sm)

			if Err.Err != nil {
				ctx.Resp.WriteHeader(http.StatusForbidden)
				return
			} else if Err.StatusCode >= 300 {
				ctx.Resp.WriteHeader(Err.StatusCode)
				return
			}

			return
		}

	case "PUT":
		if flag := strings.Contains(ctx.Req.RequestURI, "/blobs/uploads/"); flag == true {
			param := ctx.Params(":uuid")
			uuid := strings.Split(param, "?")[0]
			layer, _ := ioutil.ReadFile(fmt.Sprintf("%v/%v/layer", setting.ImagePath, uuid))
			digest := ctx.Query("digest")

			desc := Descriptor{
				MediaType: DefaultMedisType,
				Size:      int64(len(layer)),
				Digest:    digest,
			}

			req := newReqRecord(utils.EncodeBasicAuth(namespace, "putblobsv2"), ctx.Req.Request)
			requrl, err := module.NewURLBuilderFromRequest(ctx.Req.Request).BuildBlobURL(repo, desc.Digest)
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Get blobs and get request URL failed, error:: %v", err.Error())
			}

			b := newBridge(requrl, actor, req, notice)
			Err := b.createBlobEventAndWrite(EventActionPush, repo, desc)
			if Err.Err != nil {
				ctx.Resp.WriteHeader(http.StatusForbidden)
				return
			} else if Err.StatusCode >= 300 {
				ctx.Resp.WriteHeader(Err.StatusCode)
				return
			}
			return
		}

		if flag := strings.Contains(ctx.Req.RequestURI, "/manifests/"); flag == true {
			buf, _ := ctx.Req.Body().Bytes()

			handler.ManifestCtx = buf

			var sm SignedManifest
			if err := json.Unmarshal(buf, &sm); err != nil {
				fmt.Errorf("Unmarshal manifest error")
			}

			sm.Raw = make([]byte, len(buf), len(buf))
			copy(sm.Raw, buf)

			digest, err := signature.DigestManifest(buf)
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())
				//return http.StatusBadRequest, result
			}
			req := newReqRecord(utils.EncodeBasicAuth(namespace, "putmanifestv2"), ctx.Req.Request)
			requrl, err := module.NewURLBuilderFromRequest(ctx.Req.Request).BuildManifestURL(sm.Name, digest)
			if err != nil {
				fmt.Errorf("[REGISTRY API V2] Put manifest and get request URL failed, error:: %v", err.Error())
			}

			b := newBridge(requrl, actor, req, notice)
			Err := b.createManifestEventAndWrite(EventActionPush, repo, &sm)
			if Err.Err != nil {
				ctx.Resp.WriteHeader(http.StatusForbidden)
				return
			} else if Err.StatusCode >= 300 {
				ctx.Resp.WriteHeader(Err.StatusCode)
				return
			}
			return
		}

	default:
		return
	}
}
