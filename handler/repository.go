package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func PutTagV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	bodystr, _ := ctx.Req.Body().String()
	log.Debug("[REGISTRY API V1] Repository Tag : %v", bodystr)

	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(bodystr)

	repo := new(models.Repository)
	if err := repo.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		log.Error("[REGISTRY API V1] Put repository tag error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": err.Error()})
		return http.StatusBadRequest, result
	}

	return http.StatusOK, []byte("true")
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if err := r.PutImages(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Put images error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put V1 images error"})
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

	//ctx.Resp.Header().Set("X-Docker-Endpoints", setting.Domains)

	return http.StatusNoContent, []byte("")
}

func GetRepositoryImagesV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get V1 repository images failed,wrong name or repository"})
		return http.StatusBadRequest, result
	} else if has == false {
		log.Error("[REGISTRY API V1] Read repository no found, %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Get V1 repository images failed,repository no found"})
		return http.StatusNotFound, result
	}

	repo.Download += 1

	if err := repo.Save(); err != nil {
		log.Error("[REGISTRY API V1] Update download count error: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Save V1 repository failed"})
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
	//	ctx.Resp.Header().Set("X-Docker-Endpoints", setting.Domains)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(repo.JSON)))

	return http.StatusOK, []byte(repo.JSON)
}

func GetTagV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		log.Error("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get V1 tag failed,wrong name or repository"})
		return http.StatusBadRequest, result
	} else if has == false {
		log.Error("[REGISTRY API V1] Read repository no found. %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Get V1 tag failed,read repository no found"})
		return http.StatusNotFound, result
	}

	tag := map[string]string{}

	for _, value := range repo.Tags {
		t := new(models.Tag)
		if err := db.Get(t, value); err != nil {
			log.Error(fmt.Sprintf("[REGISTRY API V1]  %s/%s Tags is not exist", namespace, repository))

			result, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("%s/%s Tags is not exist", namespace, repository)})
			return http.StatusNotFound, result
		}

		tag[t.Name] = t.ImageId
	}

	result, _ := json.Marshal(tag)
	return http.StatusOK, result
}

func PutRepositoryV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	body, err := ctx.Req.Body().String()
	if err != nil {
		log.Error("[REGISTRY API V1] Get request body error: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Put V1 repository failed,request body is empty"})
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, body, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
		log.Error("[REGISTRY API V1] Put repository error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": err.Error()})
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

	//	ctx.Resp.Header().Set("X-Docker-Endpoints", setting.Domains)

	return http.StatusOK, []byte("\"\"")
}
