package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

func PutTagV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	bodystr, _ := ctx.Req.Body().String()

	rege, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := rege.FindStringSubmatch(bodystr)

	r := new(models.Repository)
	if err := r.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V1] Failed to save repository %v/%v tag: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save repository tag"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if err := r.PutImages(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Failed to save images: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save images"})
		return http.StatusBadRequest, result
	}

	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
		token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
			utils.MD5(username),
			namespace,
			repository,
			"write")

		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusNoContent, result
}

func GetRepositoryImagesV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Failed to get repository %v/%v context: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get repository context"})
		return http.StatusBadRequest, result
	} else if exists == false {
		log.Error("[REGISTRY API V1] Not found repository %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Not found repository"})
		return http.StatusNotFound, result
	}

	r.Download += 1

	if err := r.Save(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Failed to save repository %v/%v context: %v", namespace, repository, err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Failed to save repository context"})
		return http.StatusBadRequest, result
	}

	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
		utils.MD5(username),
		namespace,
		repository,
		"read")

	ctx.Resp.Header().Set("X-Docker-Token", token)
	ctx.Resp.Header().Set("WWW-Authenticate", token)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(r.JSON)))

	return http.StatusOK, []byte(r.JSON)
}

func GetTagV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if exists, err := r.Get(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Failed to get repository %v/%v context: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get repository context"})
		return http.StatusBadRequest, result
	} else if exists == false {
		log.Error("[REGISTRY API V1] Not found repository %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Not found repository"})
		return http.StatusNotFound, result
	}

	tags := map[string]string{}
	tagslist := r.GetTagslist()
	for _, tagname := range tagslist {
		t := new(models.Tag)
		if exists, err := t.Get(namespace, repository, tagname); err != nil || !exists {
			log.Error(fmt.Sprintf("[REGISTRY API V1]  %s/%s %v is not exist", namespace, repository, tagname))

			result, _ := json.Marshal(map[string]string{"message": "Tag is not exist"})
			return http.StatusNotFound, result
		}

		tags[tagname] = t.ImageId
	}

	result, _ := json.Marshal(tags)
	return http.StatusOK, result
}

func PutRepositoryV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	body, err := ctx.Req.Body().String()
	if err != nil {
		log.Error("[REGISTRY API V1] Failed to get request body: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Failed to get request body"})
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, body, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
		log.Error("[REGISTRY API V1] Failed to save repository %v/%v context: %v", namespace, repository, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save repository context"})
		return http.StatusBadRequest, result
	}

	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
			utils.MD5(username),
			namespace,
			repository,
			"write")

		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
