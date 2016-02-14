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
		log.Error("[REGISTRY API V2] Failed to get manifest digest: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get manifest digest"})
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		log.Error("[REGISTRY API V2] Failed to save repository %v/%v: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save repository"})
		return http.StatusInternalServerError, result
	}

	err, schema := module.ParseManifest(manifest, namespace, repository, tag)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to decode manifest: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to decode manifest"})
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
	if _, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V2] Failed to get repository %v/%v: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get repository"})
		return http.StatusBadRequest, result
	}

	data := map[string]interface{}{}

	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	tagslist := r.GetTagslist()
	if len(tagslist) <= 0 {
		log.Error("[REGISTRY API V2] Repository %v/%v tags list is empty", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Tags list is empty"})
		return http.StatusInternalServerError, result
	}
	data["tags"] = tagslist

	result, _ := json.Marshal(data)
	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err != nil || !exists {
		log.Error("[REGISTRY API V2] Not found manifest: %v", err)

		result, _ := json.Marshal(map[string]string{"message": "Not found manifest"})
		return http.StatusNotFound, result
	}

	digest, err := utils.DigestManifest([]byte(t.Manifest))
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to get manifest digest: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get manifest digest"})
		return http.StatusInternalServerError, result
	}

	contenttype := []string{"", "application/json; charset=utf-8", "application/vnd.docker.distribution.manifest.v2+json"}
	ctx.Resp.Header().Set("Content-Type", contenttype[t.Schema])

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}
