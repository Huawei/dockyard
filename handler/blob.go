package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Unknwon/macaron"

	"github.com/containerops/crew/modules"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
)

func authorizationVerify(ctx *macaron.Context) error {
	authinfo := ctx.Req.Header.Get("Authorization")
	if len(authinfo) == 0 {
		return fmt.Errorf("Invalid authorization")
	}

	username, passwd, err := utils.DecodeBasicAuth(authinfo)
	if err != nil {
		return err
	}

	if _, err := modules.GetUser(username, passwd); err != nil {
		return err
	}

	//TBD: verify the digest

	return nil
}

func HeadBlobsV2Handler(ctx *macaron.Context) (int, []byte) {

	if err := authorizationVerify(ctx); err != nil {
		result, _ := json.Marshal(map[string]string{"Error": "Invalid authorization"})
		return http.StatusUnauthorized, result
	}

	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]
	i := new(models.Image)
	if has, _, _ := i.HasTarsum(tarsum); has == false {
		result, _ := json.Marshal(map[string]string{"Error": "Digest not found"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")

	return http.StatusOK, []byte("")
}

func PostBlobsV2Handler(ctx *macaron.Context) (int, []byte) {

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	uuid := db.GeneralDBKey(fmt.Sprintf("%s/%s", namespace, repository))
	state := db.GeneralDBKey(fmt.Sprintf("%s/%s/%s", namespace, repository, uuid))
	random := fmt.Sprintf("http://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		"containerops.me", //TBD: code like this just for test,it will be update after config is ready
		namespace,
		repository,
		uuid,
		state)

	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	//ctx.Resp.Header().Set("Range", "0-0")

	return http.StatusAccepted, []byte("")
}

func PutBlobsV2Handler(ctx *macaron.Context) (int, []byte) {

	digest := ctx.Query("digest")
	tarsum := strings.Split(digest, ":")[1]
	imagePath := fmt.Sprintf("%v/uuid/%v", setting.BasePath, tarsum)
	layerfile := fmt.Sprintf("%v/uuid/%v/layer", setting.BasePath, tarsum)

	i := new(models.Image)
	if err := i.PutTarsum(tarsum); err != nil {
		result, _ := json.Marshal(map[string]string{"Error": "Save tarsum failure"})
		return http.StatusBadRequest, result
	}

	if !utils.IsDirExists(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		result, _ := json.Marshal(map[string]string{"Error": "Save layerfile failure"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("http://%s/v2/%s/%s/blobs/%s",
		"containerops.me", //TBD: code like this just for test,it will be update after config is ready
		ctx.Params(":namespace"),
		ctx.Params(":repository"),
		digest)

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusCreated, []byte("")
}

func GetBlobsV2Handler(ctx *macaron.Context) (int, []byte) {

	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if has, _, _ := i.HasTarsum(tarsum); has == false {
		result, _ := json.Marshal(map[string]string{"message": "Digest not found"})
		return http.StatusNotFound, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "File path is invalid"})
		return http.StatusBadRequest, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Read file failure"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
}
