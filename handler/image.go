package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/macaron"
	crew "github.com/containerops/crew/models"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
	"github.com/containerops/wrench/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func GetImageAncestryV1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read Image Ancestry Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Read Image Ancestry Error"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read Image None: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Read Image None"})
		return http.StatusBadRequest, result
	}

	//TBC
	if _, err := ctx.Resp.Write([]byte(i.Ancestry)); err != nil {
		fmt.Errorf("[REGISTRY API V1] GetImageAncestryV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func GetImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")
	var jsonInfo string
	var checksum string
	var err error

	if jsonInfo, err = models.GetJSON(imageId); err != nil {
		fmt.Errorf("[REGISTRY API V1] Search Image JSON Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Search Image JSON Error"})
		return http.StatusBadRequest, result
	}

	if checksum, err = models.GetChecksum(imageId); err != nil {
		fmt.Errorf("[REGISTRY API V1] Search Image Checksum Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Search Image Checksum Error"})
		return http.StatusBadRequest, result
	} else {
		ctx.Resp.Header().Set("X-Docker-Checksum", checksum)
	}

	//TBC
	if _, err := ctx.Resp.Write([]byte(jsonInfo)); err != nil {
		fmt.Errorf("[REGISTRY API V1] GetImageJSONV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func GetImageLayerV1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read Image Layer File Status Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Read Image Layer file Error"})
		return http.StatusBadRequest, result
	} else if has == false {
		fmt.Errorf("[REGISTRY API V1] Read Image None Error")

		result, _ := json.Marshal(map[string]string{"Error": "Read Image None"})
		return http.StatusBadRequest, result
	}

	layerfile := i.Path
	if _, err := os.Stat(layerfile); err != nil {
		fmt.Errorf("[REGISTRY API V1] Read Image file state error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Read Image file state error"})
		return http.StatusBadRequest, result
	}

	file, err := ioutil.ReadFile(layerfile)
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Read Image file error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Read Image file error"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Resp.Header().Set("Content-Length", string(int64(len(file))))

	if _, err := ctx.Resp.Write(file); err != nil {
		fmt.Errorf("[REGISTRY API V1] GetImageLayerV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func PutImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")

	jsonInfo, err := ctx.Req.Body().String()
	if err != nil {
		fmt.Errorf("[REGISTRY API V1] Get request body error: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": ""})
		return http.StatusBadRequest, result
	}

	if err := models.PutJSON(imageId, jsonInfo, setting.APIVERSION_V1); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put Image JSON Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Image JSON Error"})
		return http.StatusBadRequest, result
	}
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := image.Log(models.ACTION_PUT_IMAGES_JSON, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Error:", err.Error())
		}
	*/
	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutImageJSONV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func PutImageLayerv1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")
	basePath := setting.BasePath
	imagePath := fmt.Sprintf("%v/images/%v", basePath, imageId)
	layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, imageId)

	if !utils.IsDirExists(imagePath) {
		os.MkdirAll(imagePath, os.ModePerm)
	}

	if _, err := os.Stat(layerfile); err == nil {
		os.Remove(layerfile)
	}

	data, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put Image Layer File Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Image Layer File Error"})
		return http.StatusBadRequest, result
	}

	if err := models.PutLayer(imageId, layerfile, true, int64(len(data))); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put Image Layer File Data Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Image Layer File Data Error"})
		return http.StatusBadRequest, result
	}
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := image.Log(models.ACTION_PUT_IMAGES_LAYER, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Error:", err.Error())
		}
	*/
	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutImageLayerv1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}

func PutImageChecksumV1Handler(ctx *macaron.Context) (int, []byte) {

	imageId := ctx.Params(":image_id")

	checksum := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum"), ":")[1]
	payload := strings.Split(ctx.Req.Header.Get("X-Docker-Checksum-Payload"), ":")[1]

	fmt.Println("[REGISTRY API V1] Image Checksum : ", checksum)
	fmt.Println("[REGISTRY API V1] Image Payload: ", payload)

	if err := models.PutChecksum(imageId, checksum, true, payload); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put Image Checksum & Payload Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Image Checksum & Payload Error"})
		return http.StatusBadRequest, result
	}

	if err := models.PutAncestry(imageId); err != nil {
		fmt.Errorf("[REGISTRY API V1] Put Image Ancestry Error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"Error": "Put Image Ancestry Error"})
		return http.StatusBadRequest, result
	}
	/*
		memo, _ := json.Marshal(this.Ctx.Input.Header)
		if err := image.Log(models.ACTION_PUT_IMAGES_CHECKSUM, models.LEVELINFORMATIONAL, models.TYPE_APIV1, image.UUID, memo); err != nil {
			beego.Error("[REGISTRY API V1] Log Error:", err.Error())
		}
	*/

	if _, err := ctx.Resp.Write([]byte("")); err != nil {
		fmt.Errorf("[REGISTRY API V1] PutImageChecksumV1Handler write response content Error")
	}

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusOK, result
}
