package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/macaron"
	crew "github.com/containerops/crew/models"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
	"net/http"
	"regexp"
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
		return http.StatusForbidden, result
	}
	//TBD
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := repo.Log(models.ACTION_PUT_TAG, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
	*/
	//this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	//this.Ctx.Output.Context.Output.Body([]byte(""))
	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutTagV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": "Put repository tag success"})
	return http.StatusOK, result
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if err := r.PutImages(namespace, repository, ctx); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put images error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put images error"})
		return http.StatusBadRequest, result
	}
	//TBD
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := repo.Log(models.ACTION_PUT_REPO_IMAGES, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
	*/

	//this.Ctx.Output.Context.Output.Body([]byte(""))
	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutRepositoryImagesV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusNoContent, result
}

func GetRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read repository json error"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read repository no found, %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Read repository no found"})
		return http.StatusBadRequest, result
	}

	repo.Download += 1

	if err := repo.Save(); err != nil {
		fmt.Errorf("[REGISTRY API V1] Update download count error: %v", err.Error())
	}
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := repo.Log(models.ACTION_GET_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
	*/
	if _, err := ctx.Resp.Write([]byte(repo.JSON)); err != nil {
		fmt.Errorf("[REGISTRY API V1] GetRepositoryImagesV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func GetTagV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	repo := new(models.Repository)
	if has, _, err := repo.Has(namespace, repository); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read repository json error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read repository json error"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read repository no found. %v/%v", namespace, repository)

		result, _ := json.Marshal(map[string]string{"message": "Read repository no found"})
		return http.StatusBadRequest, result
	}

	tag := map[string]string{}

	for _, value := range repo.Tags {
		t := new(models.Tag)
		if err := db.Get(t, value); err != nil {
			fmt.Errorf(fmt.Sprintf("[REGISTRY API V1]  %s/%s Tags is not exist", namespace, repository))

			result, _ := json.Marshal(map[string]string{"message": fmt.Sprintf("%s/%s Tags is not exist", namespace, repository)})
			return http.StatusBadRequest, result
		}

		tag[t.Name] = t.ImageId
	}

	//this.Data["json"] = tag
	//this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	//this.ServeJson()

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func PutRepositoryV1Handler(ctx *macaron.Context) (int, []byte) {

	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	requestbody, err := ctx.Req.Body().String()
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Get request body error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": ""})
		return http.StatusForbidden, result
	}

	r := new(models.Repository)
	if err := r.Put(namespace, repository, requestbody, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put repository error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		//TBD : code as below just for testing,it will be updated later
		//return http.StatusForbidden, result
		return http.StatusOK, result
	}

	if ctx.Req.Header.Get("X-Docker-Token") == "true" {
		token := db.GeneralDBKey(username)
		//this.SetSession("token", token)
		ctx.Resp.Header().Set("X-Docker-Token", token)
		ctx.Resp.Header().Set("WWW-Authenticate", token)
	}

	//memo, _ := json.Marshal(this.Ctx.Input.Header)

	user := new(crew.User)
	if _, _, err := user.Has(username); err != nil {
		fmt.Errorf("[REGISTRY API V1] Get user error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusForbidden, result
	}
	//TBD
	/*
		if err := user.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
		if err := repo.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
	*/

	//TBD:Endpoints should be read from APP configfile
	ctx.Resp.Header().Set("X-Docker-Endpoints", "containerops.com")

	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutRepositoryV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}
