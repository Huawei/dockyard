package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func PutManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	agent := ctx.Req.Header.Get("User-Agent")

	repo := new(models.Repository)
	if err := repo.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		log.Error("[REGISTRY API V2] Save repository failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": err.Error()})
		return http.StatusBadRequest, result
	}

	manifest, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := models.ManifestconvertV1(manifest); err != nil {
		log.Error("[REGISTRY API V2] Decode Manifest Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Manifest converted failed"})
		return http.StatusBadRequest, result
	}

	digest, err := utils.DigestManifest(manifest)
	if err != nil {
		log.Error("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get manifest digest failed"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("https://%v/v2/%v/%v/manifests/%v",
		setting.Domains,
		namespace,
		repository,
		digest)

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusAccepted, []byte("")
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
	t := new(models.Tag)

	if err := t.Get(ctx.Params(":namespace"), ctx.Params(":repository"), ctx.Params(":tag")); err != nil {
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

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}
