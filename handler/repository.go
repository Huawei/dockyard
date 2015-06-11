package handler

import (
	"encoding/json"
	"net/http"

	"fmt"
	"github.com/Unknwon/macaron"
	"github.com/containerops/crew/models"
	"github.com/containerops/crew/setting"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
)

func PutTagV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func PutRepositoryImagesV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
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
	if true {
		username, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))

		namespace := string(ctx.Params(":namespace"))
		repository := string(ctx.Params(":repository"))

		requestbody, err := ctx.Req.Body().String()
		if err != nil {
			fmt.Println("[REGISTRY API V1] Get request body error:", err.Error())

			result, _ := json.Marshal(map[string]string{"message": ""})
			return http.StatusForbidden, result
		}

		repo := new(models.Repository)
		if err := repo.Put(namespace, repository, requestbody, ctx.Req.Header.Get("User-Agent"), setting.APIVERSION_V1); err != nil {
			fmt.Println("[REGISTRY API V1] Put repository error:", err.Error())

			result, _ := json.Marshal(map[string]string{"Error": err.Error()})
			//TBD : code as below just for testing,it will be updated later
			//return http.StatusForbidden, result
			return http.StatusOK, result
		}

		if ctx.Req.Header.Get("X-Docker-Token") == "true" {
			token := db.GeneralDBKey(username)
			//this.SetSession("token", token)
			ctx.Resp.Header().Set("X-Docker-Token", token)
			//ctx.Resp.Header().Set("WWW-Authenticate", token)
		}

		//memo, _ := json.Marshal(this.Ctx.Input.Header)

		user := new(models.User)
		if _, _, err := user.Has(username); err != nil {
			fmt.Println("[REGISTRY API V1] Get user error:", err.Error())

			result, _ := json.Marshal(map[string]string{"Error": err.Error()})
			return http.StatusForbidden, result
		}
		/*
			if err := user.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
				beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
			}
			if err := repo.Log(models.ACTION_UPDATE_REPO, models.LEVELINFORMATIONAL, models.TYPE_APIV1, repo.UUID, memo); err != nil {
				beego.Error("[REGISTRY API V1] Log Erro:", err.Error())
			}
		*/
		//TBD:Endpoints should be read from APP configfile
		ctx.Resp.Header().Set("X-Docker-Endpoints", "dockyard.com")

		//TBD:
		//this.Ctx.Output.Context.Output.Body([]byte(""))

		result, _ := json.Marshal(map[string]string{"message": ""})
		return http.StatusOK, result
	} else {

		return 404, nil
	}
}
