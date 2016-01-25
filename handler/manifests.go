package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

var ManifestCtx []byte

func PutManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//TODO: to consider parallel situation
	manifest := ManifestCtx
	defer func() {
		ManifestCtx = []byte{}
	}()

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	agent := ctx.Req.Header.Get("User-Agent")
	tag := ctx.Params(":tag")

	if len(manifest) == 0 {
		manifest, _ = ctx.Req.Body().Bytes()
	}

	digest, err := utils.DigestManifest(manifest)
	if err != nil {
		log.Error("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get manifest digest failed"})
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		log.Error("[REGISTRY API V2] Save repository failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": err.Error()})
		return http.StatusBadRequest, result
	}

	err, schema := module.ParseManifest(manifest, namespace, repository, tag)
	if err != nil {
		log.Error("[REGISTRY API V2] Decode Manifest Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Decode Manifest Error"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/%s/manifests/%s",
		setting.ListenMode,
		setting.Domains,
		namespace,
		repository,
		digest)

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	var status = []int{http.StatusBadRequest, http.StatusAccepted, http.StatusCreated}
	result, _ := json.Marshal(map[string]string{})
	return status[schema], result
}

func GetTagsListV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if has, _, err := r.Has(namespace, repository); err != nil || has == false {
		log.Error("[REGISTRY API V2] Repository not found: %v", repository)

		result, _ := json.Marshal(map[string]string{"message": "Repository not found"})
		return http.StatusNotFound, result
	}

	data := map[string]interface{}{}
	tags := []string{}

	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	for _, value := range r.Tags {
		t := new(models.Tag)
		if err := t.GetByKey(value); err != nil {
			log.Error("[REGISTRY API V2] Tag not found: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Tag not found"})
			return http.StatusNotFound, result
		}

		tags = append(tags, t.Name)
	}

	data["tags"] = tags

	result, _ := json.Marshal(data)
	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	t := new(models.Tag)
	if err := t.Get(namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V2] Manifest not found: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Manifest not found"})
		return http.StatusNotFound, result
	}

	digest, err := utils.DigestManifest([]byte(t.Manifest))
	if err != nil {
		log.Error("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get manifest digest failed"})
		return http.StatusBadRequest, result
	}

	contenttype := []string{"", "application/json; charset=utf-8", "application/vnd.docker.distribution.manifest.v2+json"}
	ctx.Resp.Header().Set("Content-Type", contenttype[t.Schema])

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}
