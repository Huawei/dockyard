package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/satori/go.uuid"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

//Support to fetch acis
func GetPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	pubkeysPath := module.GetPubkeysPath(namespace, repository)
	if _, err := os.Stat(pubkeysPath); err != nil {
		log.Error("[ACI API] Pubkeys path %v is invalid: %v", pubkeysPath, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Pubkeys path is invalid"})
		return http.StatusInternalServerError, result
	}

	files, err := ioutil.ReadDir(pubkeysPath)
	if err != nil {
		log.Error("[ACI API] Failed to read pubkeys path %v: %v", pubkeysPath, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to read pubkeys path"})
		return http.StatusInternalServerError, result
	}

	var pubkey []byte
	if len(files) <= 0 {
		log.Error("[ACI API] Not found pubkey")

		result, _ := json.Marshal(map[string]string{"message": "Not found pubkey"})
		return http.StatusNotFound, result
	}

	// TODO: support single pubkey per user now, to consider whether to support multiple pubkeys per user in the future
	filename := pubkeysPath + "/" + files[0].Name()
	pubkey, err = ioutil.ReadFile(filename)
	if err != nil {
		log.Error("[ACI API] Failed to read pubkey %v: %v", filename, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to read pubkey"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	acifilename := ctx.Params(":acifilename")

	acifile := strings.Trim(acifilename, ".asc")
	tag := strings.Trim(acifile, ".aci")

	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err != nil || !exists {
		log.Error("[ACI API] Not found ACI %v/%v/%v", namespace, repository, acifilename)
		result, _ := json.Marshal(map[string]string{"message": "Not found ACI"})
		return http.StatusNotFound, result
	}

	i := new(models.Image)
	if exists, err := i.Get(t.ImageId); err != nil || !exists {
		log.Error("[ACI API] Not found ACI %v/%v/%v", namespace, repository, acifilename)
		result, _ := json.Marshal(map[string]string{"message": "Not found ACI"})
		return http.StatusNotFound, result
	}

	var filepath string
	if b := strings.Contains(acifilename, ".asc"); b == true {
		filepath = i.SignPath
	} else {
		filepath = i.AciPath
	}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Error("[ACI API] Failed to get ACI file %v: %v", filepath, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to get ACI file"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, content
}

//Support to push acis
func PostUploadHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	acifile := ctx.Params(":acifile")
	signfile := fmt.Sprintf("%v%v", acifile, ".asc")

	//TODO: only for testing,pubkey will be read and saved via user management module
	pubkeyspath := module.GetPubkeysPath(namespace, repository)
	if _, err := os.Stat(pubkeyspath); err != nil {
		if err := os.MkdirAll(pubkeyspath, os.ModePerm); err != nil {
			log.Error("[ACI API] Failed to create pubkeys path %v: %v", pubkeyspath, err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Failed to create pubkeys path"})
			return http.StatusInternalServerError, result
		}
	}

	imageId := utils.MD5(uuid.NewV4().String())
	imagepath := module.GetImagePath(imageId)
	if err := os.MkdirAll(imagepath, os.ModePerm); err != nil {
		log.Error("[ACI API] Failed to create aci path %v: %v", imagepath, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Failed to create aci path"})
		return http.StatusInternalServerError, result
	}

	prefix := fmt.Sprintf("%v://%v/ac/push/%v/%v/", setting.ListenMode, setting.Domains, namespace, repository)
	endpoint := models.UploadDetails{
		ACIPushVersion: "0.0.1", //TODO: follow ACI push spec
		Multipart:      false,
		ManifestURL:    prefix + imageId + "/manifest",
		SignatureURL:   prefix + imageId + "/signature/" + signfile,
		ACIURL:         prefix + imageId + "/aci/" + acifile,
		CompletedURL:   prefix + imageId + "/complete/" + acifile + "/" + signfile,
	}

	result, _ := json.Marshal(endpoint)
	return http.StatusOK, result
}

func PutManifestHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")

	manipath := module.GetManifestPath(imageId)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(manipath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Failed to save manifest file: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Failed to save manifest"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")
	signfile := ctx.Params(":signfile")

	signpath := module.GetSignaturePath(imageId, signfile)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signpath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Failed to save signature file %v : %v", signpath, err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Failed to save signature file"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")
	acifile := ctx.Params(":acifile")

	acipath := module.GetAciPath(imageId, acifile)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(acipath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Failed to save aci file %v : %v", acipath, err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Failed to save aci file"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostCompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	imageId := ctx.Params(":imageId")
	repository := ctx.Params(":repository")

	body, _ := ctx.Req.Body().Bytes()
	if err := module.CheckClientStatus(body); err != nil {
		module.CleanCache(imageId)
		log.Error("[ACI API] Failed to push aci: %v", err.Error())

		failmsg := module.FillRespMsg(false, err.Error(), "")
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	namespace := ctx.Params(":namespace")
	acifile := ctx.Params(":acifile")
	signfile := ctx.Params(":signfile")

	//TODO: only for testing,pubkey will be read and saved via user management module
	pubkeyspath := module.GetPubkeysPath(namespace, repository)
	acipath := module.GetAciPath(imageId, acifile)
	signpath := module.GetSignaturePath(imageId, signfile)
	manipath := module.GetManifestPath(imageId)
	if err := module.VerifyAciSignature(acipath, signpath, pubkeyspath); err != nil {
		module.CleanCache(imageId)
		log.Error("[ACI API] Failed to verify Aci: %v", err.Error())

		failmsg := module.FillRespMsg(false, "", err.Error())
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	//If aci image is existent,it should update the db and delete the old image after executed successfully
	var oldimageId = ""
	tag := strings.Trim(acifile, ".aci")
	t := new(models.Tag)
	if exists, err := t.Get(namespace, repository, tag); err == nil && exists {
		oldimageId = t.ImageId
	}

	a := new(models.Aci)
	if err := a.Update(namespace, repository, tag, imageId, manipath, signpath, acipath); err != nil {
		module.CleanCache(imageId)
		log.Error("[ACI API] Failed to update %v/%v: %v", namespace, repository, err.Error())

		failmsg := module.FillRespMsg(false, "", err.Error())
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	//Delete old aci directory after redis is updated
	if oldimageId != "" {
		module.CleanCache(oldimageId)
	}

	successmsg := module.FillRespMsg(true, "", "")
	result, _ := json.Marshal(successmsg)
	return http.StatusOK, result
}
