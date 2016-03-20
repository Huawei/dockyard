package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

func GetImageAncestryV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	i := new(models.Image)
	if _, err := i.Get(imageId); err != nil {
		log.Error("[REGISTRY API V1] Failed to get image %v ancestry: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get image ancestry"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(i.Ancestry)))

	return http.StatusOK, []byte(i.Ancestry)
}

func GetImageJSONV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var jsonInfo string
	var payload string
	var err error

	imageId := ctx.Params(":imageId")

	i := new(models.Image)
	if jsonInfo, err = i.GetJSON(imageId); err != nil {
		log.Error("[REGISTRY API V1] Failed to get image %v json: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get image json"})
		return http.StatusNotFound, result
	}

	if payload, err = i.GetPayload(imageId); err != nil {
		log.Error("[REGISTRY API V1] Failed to get image %v payload: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get image payload"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("X-Docker-Checksum-Payload", fmt.Sprintf("sha256:%v", payload))
	ctx.Resp.Header().Set("X-Docker-Size", fmt.Sprint(i.Size))
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(jsonInfo)))

	return http.StatusOK, []byte(jsonInfo)
}

func GetImageLayerV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	i := new(models.Image)
	if _, err := i.Get(imageId); err != nil {
		log.Error("[REGISTRY API V1] Failed to get image %v layer: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get image layer"})
		return http.StatusNotFound, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		log.Error("[REGISTRY API V1] Image layer file state error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Image layer file state error"})
		return http.StatusInternalServerError, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		log.Error("[REGISTRY API V1] Failed to read image layer: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to read image layer"})
		return http.StatusInternalServerError, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
}

func PutImageJSONV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	info, err := ctx.Req.Body().String()
	if err != nil {
		log.Error("[REGISTRY API V1] Failed to get request body: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get request body"})
		return http.StatusBadRequest, result
	}

	i := new(models.Image)
	if err := i.PutJSON(imageId, info, setting.APIVERSION_V1); err != nil {
		log.Error("[REGISTRY API V1] Failed to put image %v json: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to put image json"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutImageLayerv1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	basePath := setting.ImagePath
	imagePath := fmt.Sprintf("%v/images/%v", basePath, imageId)
	layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

	if !utils.IsDirExist(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		log.Error("[REGISTRY API V1] Failed to save image layer: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save image layer"})
		return http.StatusBadRequest, result
	}

	i := new(models.Image)
	if err := i.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		log.Error("[REGISTRY API V1] Failed to save image %v layer data: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save image layer data"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutImageChecksumV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	checksum := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum"), ":")[1]
	payload := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum-Payload"), ":")[1]

	i := new(models.Image)
	if err := i.PutChecksum(imageId, checksum, true, payload); err != nil {
		log.Error("[REGISTRY API V1] Failed to save image %v checksum: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save image checksum"})
		return http.StatusBadRequest, result
	}

	if err := i.PutAncestry(imageId); err != nil {
		log.Error("[REGISTRY API V1] Failed to save image %v ancestry: %v", imageId, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to save image ancestry"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}
