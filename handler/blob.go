package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
	"github.com/satori/go.uuid"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func HeadBlobsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	digest := ctx.Params(":digest")
	tarsum := strings.Split(digest, ":")[1]

	i := new(models.Image)
	if has, _ := i.HasTarsum(tarsum); has == false {
		log.Info("[REGISTRY API V2] Tarsum not found: %v", tarsum)

		result, _ := json.Marshal(map[string]string{"message": "Tarsum not found"})
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
	state := utils.MD5(fmt.Sprintf("%s/%s/%s", namespace, repository, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
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

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfileTmp, data, 0777); err != nil {
		log.Error("[REGISTRY API V2] Save layerfile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save layerfile failed"})
		return http.StatusBadRequest, result
	}

	state := utils.MD5(fmt.Sprintf("%s/%s/%s", namespace, repository, time.Now().UnixNano()/int64(time.Millisecond)))
	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/uploads/%s?_state=%s",
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
	layerlen, err := modules.CopyImgLayer(imagePathTmp, layerfileTmp, imagePath, layerfile, ctx.Req.Request.Body)
	if err != nil {
		log.Error("[REGISTRY API V2] Save layerfile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save layerfile failed"})
		return http.StatusBadRequest, result
	}

	//saving specific tarsum every times is in order to split the same tarsum in HEAD handler
	i := new(models.Image)
	i.Path, i.Size = layerfile, int64(layerlen)
	if err := i.PutTarsum(tarsum); err != nil {
		log.Error("[REGISTRY API V2] Save tarsum failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save tarsum failed"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("https://%s/v2/%s/%s/blobs/%s",
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
	has, _ := i.HasTarsum(tarsum)
	if has == false {
		log.Error("[REGISTRY API V2] Digest not found: %v", tarsum)

		result, _ := json.Marshal(map[string]string{"message": "Digest not found"})
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
