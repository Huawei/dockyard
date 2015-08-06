package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func authorizationVerify(ctx *macaron.Context) error {
	return nil
}

func HeadBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	if err := authorizationVerify(ctx); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Invalid authorization"})
		return http.StatusUnauthorized, result
	}

	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]
	i := new(models.Image)
	if has, _, _ := i.HasTarsum(tarsum); has == false {
		result, _ := json.Marshal(map[string]string{"message": "Digest not found"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")

	return http.StatusOK, []byte("")
}

func PostBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	key := db.Key("repository", namespace, repository)
	state := utils.MD5(fmt.Sprintf("%s/%s/%s", namespace, repository, key))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		"containerops.me", //TBD: code like this just for test,it will be update after config is ready
		namespace,
		repository,
		key,
		state)

	ctx.Resp.Header().Set("Docker-Upload-Uuid", key)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusAccepted, []byte("")
}

func PutBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	digest := ctx.Query("digest")
	tarsum := strings.Split(digest, ":")[1]
	imagePathTmp := fmt.Sprintf("%v/temp/%v", setting.ImagePath, tarsum)
	layerfileTmp := fmt.Sprintf("%v/temp/%v/layer", setting.ImagePath, tarsum)

	i := new(models.Image)
	if err := i.PutTarsum(tarsum); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Save tarsum failure"})
		return http.StatusBadRequest, result
	}

	if !utils.IsDirExists(imagePathTmp) {
		os.MkdirAll(imagePathTmp, os.ModePerm)
	}

	if _, err := os.Stat(layerfileTmp); err == nil {
		os.Remove(layerfileTmp)
	}

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfileTmp, data, 0777); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Save layerfile failure"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/%s",
		"containerops.me", //TBD: code like this just for test,it will be update after config is ready
		ctx.Params(":namespace"),
		ctx.Params(":repository"),
		digest)

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusCreated, []byte("")
}

func GetBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

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
