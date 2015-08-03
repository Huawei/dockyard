package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/Unknwon/macaron"

	crew "github.com/containerops/crew/models"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
)

func PutTagV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	tag := ctx.Params(":tag")

	bodystr, _ := ctx.Req.Body().String()
	fmt.Println("[REGISTRY API V1] Repository Tag:", bodystr)

	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(bodystr)

	repo := new(models.Repository)
	if err := repo.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put repository tag error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusBadRequest, result
	}

	return http.StatusOK, []byte("true")
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if err := r.PutImages(namespace, repository, ctx); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put images error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put V1 images error"})
		return http.StatusBadRequest, result
	}

	//TBD: set content like docker format?
	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
		token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
			db.GeneralDBKey(username),
			namespace,
			repository,
			"write")
		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	//TBD: the head value will be got from config
	ctx.Resp.Header().Set("X-Docker-Endpoints", "containerops.me")

	return http.StatusNoContent, []byte("")
}

func GetRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get V1 repository images failure,wrong name or repository"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read repository no found, %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Get V1 repository images failure,repository no found"})
		return http.StatusNotFound, result
	}

	repo.Download += 1

	if err := repo.Save(); err != nil {
		fmt.Errorf("[REGISTRY API V1] Update download count error: %v", err.Error())
	}

	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
		db.GeneralDBKey(username),
		namespace,
		repository,
		"read")
	ctx.Resp.Header().Set("X-Docker-Token", token)
	ctx.Resp.Header().Set("WWW-Authenticate", token)
	ctx.Resp.Header().Set("X-Docker-Endpoints", "containerops.me")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(repo.JSON)))

	return http.StatusOK, []byte(repo.JSON)
}

func GetTagV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get V1 tag failure,wrong name or repository"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read repository no found. %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Get V1 tag failure,read repository no found"})
		return http.StatusNotFound, result
	}

	tag := map[string]string{}

	for _, value := range repo.Tags {
		t := new(models.Tag)
		if err := db.Get(t, value); err != nil {
			fmt.Errorf(fmt.Sprintf("[REGISTRY API V1]  %s/%s Tags is not exist", namespace, repository))

			result, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("%s/%s Tags is not exist", namespace, repository)})
			return http.StatusNotFound, result
		}

		tag[t.Name] = t.ImageId
	}

	result, _ := json.Marshal(tag)
	return http.StatusOK, result
}

func PutRepositoryV1Handler(ctx *macaron.Context) (int, []byte) {

	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	requestbody, err := ctx.Req.Body().String()
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Get request body error: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Put V1 repository failure,request body is empty"})
		return http.StatusBadRequest, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, requestbody, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put repository error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusBadRequest, result
	}

	//TBD: return docker format?
	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		token := fmt.Sprintf("Token signature=%v,repository=\"%v/%v\",access=%v",
			db.GeneralDBKey(username),
			namespace,
			repository,
			"write")
		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	//memo, _ := json.Marshal(this.Ctx.Input.Header)

	user := new(crew.User)
	if _, _, err := user.Has(username); err != nil {
		fmt.Errorf("[REGISTRY API V1] Get user error: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusNotFound, result
	}

	//TBD:Endpoints should be read from APP configfile
	ctx.Resp.Header().Set("X-Docker-Endpoints", "containerops.me")

	return http.StatusOK, []byte("\"\"")
}
