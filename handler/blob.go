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

func HeadBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if has, _, _ := i.HasTarsum(tarsum); has == false {
		log.Info("[REGISTRY API V2] Tarsum not found: %v", tarsum)

		result, _ := json.Marshal(map[string]string{"message": "Tarsum not found"})
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
		setting.Domains,
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

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	i := new(models.Image)
	if err := i.PutTarsum("", tarsum); err != nil {
		log.Error("[REGISTRY API V2] Save tarsum failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save tarsum failed"})
		return http.StatusBadRequest, result
	}

	if !utils.IsDirExist(imagePathTmp) {
		os.MkdirAll(imagePathTmp, os.ModePerm)
	}

	if _, err := os.Stat(layerfileTmp); err == nil {
		os.Remove(layerfileTmp)
	}

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfileTmp, data, 0777); err != nil {
		log.Error("[REGISTRY API V2] Save layerfile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save layerfile failed"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/%s",
		setting.Domains,
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
	has, imagegrp, _ := i.HasTarsum(tarsum)
	if has == false {
		log.Error("[REGISTRY API V2] Digest not found: %v", tarsum)

		result, _ := json.Marshal(map[string]string{"message": "Digest not found"})
		return http.StatusNotFound, result
	}

	if has, _, err := i.Has(imagegrp[1]); err != nil {
		log.Error("[REGISTRY API V2] Read Image Ancestry Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image Ancestry Error"})
		return http.StatusBadRequest, result
	} else if has == false {
		log.Error("[REGISTRY API V2] Read Image None: %v", imagegrp[1])

		result, _ := json.Marshal(map[string]string{"message": "Read Image None"})
		return http.StatusNotFound, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		log.Error("[REGISTRY API V2] File path is invalid: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "File path is invalid"})
		return http.StatusBadRequest, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		log.Error("[REGISTRY API V2] Read file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read file failed"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
}
