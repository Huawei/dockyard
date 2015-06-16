package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/macaron"
	"github.com/containerops/crew/models"
	"github.com/containerops/crew/setting"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
	"net/http"
	"regexp"
)

func PutTagV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repo_name")
	tag := ctx.Params(":tag")

	bodystr, _ := ctx.Req.Body().String()
	fmt.Println("[REGISTRY API V1] Repository Tag:", bodystr)

	r, _ := regexp.Compile(`"([[:alnum:]]+)"`)
	imageIds := r.FindStringSubmatch(bodystr)

	repo := new(models.Repository)
	if err := repo.PutTag(imageIds[1], namespace, repository, tag); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put repository tag error:", err.Error())

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
		fmt.Errorf("[REGISTRY API V1] PutRepositoryImagesV1Handler Write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": "Put repository tag success"})
	return http.StatusOK, result
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repo_name")

	repo := new(models.Repository)
	if err := repo.PutImages(namespace, repository); err != nil {
		fmt.Errorf("[REGISTRY API V1] Update Uploaded flag error: %s", namespace, repository, err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Update Uploaded flag error"})
		return http.StatusBadRequest, result
	}
	//TBD
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := repo.Log(models.ACTION_PUT_REPO_IMAGES, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
		}
	*/
	org := new(models.Organization)
	isOrg, _, err := org.Has(namespace)
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Search Organization Error: ", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Search Organization Error"})
		return http.StatusBadRequest, result
	}

	user := new(models.User)
	authUsername, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	isUser, _, err := user.Has(authUsername)
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Search User Error: ", err.Error())
		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusBadRequest, result
	}

	if !isUser && !isOrg {
		fmt.Errorf("[REGISTRY API V1] Search Namespace Error")
		result, _ := json.Marshal(map[string]string{"message": "Search Namespace Error"})
		return http.StatusBadRequest, result
	}

	if isUser {
		user.Repositories = append(user.Repositories, repo.UUID)
		user.Save()
	}
	if isOrg {
		org.Repositories = append(org.Repositories, repo.UUID)
		org.Save()
	}

	//this.Ctx.Output.Context.Output.SetStatus(http.StatusNoContent)
	//this.Ctx.Output.Context.Output.Body([]byte(""))
	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutRepositoryImagesV1Handler Write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusNoContent, result
}

func GetRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetTagV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func PutRepositoryV1Handler(ctx *macaron.Context) (int, []byte) {

	username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	requestbody, err := ctx.Req.Body().String()
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Get request body error:", err.Error())

		result, _ := json.Marshal(map[string]string{"message": ""})
		return http.StatusForbidden, result
	}

	repo := new(models.Repository)
	if err := repo.Put(namespace, repository, requestbody, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put repository error:", err.Error())

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

	user := new(models.User)
	if _, _, err := user.Has(username); err != nil {
		fmt.Errorf("[REGISTRY API V1] Get user error:", err.Error())

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
		fmt.Errorf("[REGISTRY API V1] PutRepositoryV1Handler Write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}
