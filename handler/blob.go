package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/satori/go.uuid"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

func HeadBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		log.Info("[REGISTRY API V2] Failed to get tarsum %v: %v", tarsum, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get tarsum"})
		return http.StatusBadRequest, result
	} else if !exists {
		log.Info("[REGISTRY API V2] Not found tarsum: %v", tarsum)

		result, _ := json.Marshal(map[string]string{"message": "Not found tarsum"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(i.Size))

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	uuid := utils.MD5(uuid.NewV4().String())
	state := utils.MD5("%s/%s/%d", namespace, repository, time.Now().UnixNano()/int64(time.Millisecond))
	random := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.ListenMode,
		setting.Domains,
		namespace,
		repository,
		uuid,
		state)

	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", "0-0")

	result, _ := json.Marshal(map[string]string{})
	return http.StatusAccepted, result
}

func PatchBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	imagePathTmp := fmt.Sprintf("%v/%v", setting.ImagePath, uuid)
	layerfileTmp := fmt.Sprintf("%v/%v/layer", setting.ImagePath, uuid)

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	if !utils.IsDirExist(imagePathTmp) {
		os.MkdirAll(imagePathTmp, os.ModePerm)
	}

	if _, err := os.Stat(layerfileTmp); err == nil {
		os.Remove(layerfileTmp)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(layerfileTmp, data, 0777); err != nil {
		log.Error("[REGISTRY API V2] Failed to save layer file %v: %v", layerfileTmp, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save layer file"})
		return http.StatusInternalServerError, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%s/%d", namespace, repository, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
		setting.ListenMode,
		setting.Domains,
		namespace,
		repository,
		uuid,
		state)

	ctx.Resp.Header().Set("Docker-Upload-Uuid", uuid)
	ctx.Resp.Header().Set("Location", random)
	ctx.Resp.Header().Set("Range", fmt.Sprintf("0-%v", len(data)-1))

	result, _ := json.Marshal(map[string]string{})
	return http.StatusAccepted, result
}

func PutBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	desc := ctx.Params(":uuid")
	uuid := strings.Split(desc, "?")[0]

	digest := ctx.Query("digest")
	tarsum := strings.Split(digest, ":")[1]

	imagePathTmp := fmt.Sprintf("%v/%v", setting.ImagePath, uuid)
	layerfileTmp := fmt.Sprintf("%v/%v/layer", setting.ImagePath, uuid)
	imagePath := fmt.Sprintf("%v/tarsum/%v", setting.ImagePath, tarsum)
	layerfile := fmt.Sprintf("%v/tarsum/%v/layer", setting.ImagePath, tarsum)

	reqbody, _ := ctx.Req.Body().Bytes()
	layerlen, err := module.CopyImgLayer(imagePathTmp, layerfileTmp, imagePath, layerfile, reqbody)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to save layer file %v: %v", layerfile, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save layer file"})
		return http.StatusInternalServerError, result
	}

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	i := new(models.Image)
	i.Path, i.Size = layerfile, int64(layerlen)
	if err := i.Save(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Failed to save tarsum %v: %v", tarsum, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save tarsum"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("%s://%s/v2/%s/%s/blobs/%s",
		setting.ListenMode,
		setting.Domains,
		ctx.Params(":namespace"),
		ctx.Params(":repository"),
		digest)

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	result, _ := json.Marshal(map[string]string{})
	return http.StatusCreated, result
}

func GetBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")

	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if exists, err := i.Get(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Failed to get tarsum %v: %v", tarsum, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get tarsum"})
		return http.StatusBadRequest, result
	} else if !exists {
		log.Error("[REGISTRY API V2] Not found tarsum: %v: %v", tarsum, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Not found tarsum"})
		return http.StatusNotFound, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		log.Error("[REGISTRY API V2] File path %v is invalid: %v", layerfile, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "File path is invalid"})
		return http.StatusInternalServerError, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		log.Error("[REGISTRY API V2] Failed to read layer file %v: %v", layerfile, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to read layer file"})
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/x-gzip")
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
}
