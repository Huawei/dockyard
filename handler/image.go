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
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func GetImageAncestryV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	imageId := ctx.Params(":imageId")

	i := new(models.Image)
	if has, _, err := i.Has(imageId); err != nil {
		log.Error("[REGISTRY API V1] Read Image Ancestry Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image Ancestry Error"})
		return http.StatusBadRequest, result
	} else if has == false {
		log.Error("[REGISTRY API V1] Read Image None: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image None"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(i.Ancestry)))
	return http.StatusOK, []byte(i.Ancestry)
}

func GetImageJSONV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	imageId := ctx.Params(":imageId")
	var jsonInfo string
	var payload string
	var err error

	i := new(models.Image)
	if jsonInfo, err = i.GetJSON(imageId); err != nil {
		log.Error("[REGISTRY API V1] Search Image JSON Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Search Image JSON Error"})
		return http.StatusNotFound, result
	}

	if payload, err = i.GetChecksumPayload(imageId); err != nil {
		log.Error("[REGISTRY API V1] Search Image Checksum Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Search Image Checksum Error"})
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
	if has, _, err := i.Has(imageId); err != nil {
		log.Error("[REGISTRY API V1] Read Image Layer File Status Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image Layer file Error"})
		return http.StatusBadRequest, result
	} else if has == false {
		log.Error("[REGISTRY API V1] Read Image None Error")

		result, _ := json.Marshal(map[string]string{"message": "Read Image None"})
		return http.StatusNotFound, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		log.Error("[REGISTRY API V1] Read Image file state error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image file state error"})
		return http.StatusBadRequest, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		log.Error("[REGISTRY API V1] Read Image file error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read Image file error"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(file)))

	return http.StatusOK, file
}

func PutImageJSONV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	imageId := ctx.Params(":imageId")

	jsonInfo, err := ctx.Req.Body().String()
	if err != nil {
		log.Error("[REGISTRY API V1] Get request body error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put V1 image JSON failed,request body is empty"})
		return http.StatusBadRequest, result
	}

	i := new(models.Image)
	if err := i.PutJSON(imageId, jsonInfo, setting.APIVERSION_V1); err != nil {
		log.Error("[REGISTRY API V1] Put Image JSON Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put Image JSON Error"})
		return http.StatusBadRequest, result
	}

	return http.StatusOK, []byte("true")
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

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		log.Error("[REGISTRY API V1] Put Image Layer File Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put Image Layer File Error"})
		return http.StatusBadRequest, result
	}

	i := new(models.Image)
	if err := i.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		log.Error("[REGISTRY API V1] Put Image Layer File Data Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put Image Layer File Data Error"})
		return http.StatusBadRequest, result
	}

	return http.StatusOK, []byte("true")
}

func PutImageChecksumV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	imageId := ctx.Params(":imageId")

	checksum := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum"), ":")[1]
	payload := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum-Payload"), ":")[1]

	log.Debug("[REGISTRY API V1] Image Checksum : %v", checksum)
	log.Debug("[REGISTRY API V1] Image Payload: %v", payload)

	i := new(models.Image)
	if err := i.PutChecksum(imageId, checksum, true, payload); err != nil {
		log.Error("[REGISTRY API V1] Put Image Checksum & Payload Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put Image Checksum & Payload Error"})
		return http.StatusBadRequest, result
	}

	if err := i.PutAncestry(imageId); err != nil {
		log.Error("[REGISTRY API V1] Put Image Ancestry Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Put Image Ancestry Error"})
		return http.StatusBadRequest, result
	}

	return http.StatusOK, []byte("true")
}
