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

func GetPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	pubkeysPath := fmt.Sprintf("%v/acis/pubkeys/%v", setting.ImagePath, namespace)
	if _, err := os.Stat(pubkeysPath); err != nil {
		log.Error("[ACI API] Search pubkeys path failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Search pubkeys path failed"})
		return http.StatusInternalServerError, result
	}

	files, err := ioutil.ReadDir(pubkeysPath)
	if err != nil {
		log.Error("[ACI API] Get pubkeys file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get pubkeys file failed"})
		return http.StatusInternalServerError, result
	}

	// TODO: considering that one user has multiple pubkeys
	var pubkey []byte
	if len(files) <= 0 {
		log.Error("[ACI API] Not found pubkey")

		result, _ := json.Marshal(map[string]string{"message": "Not found pubkey"})
		return http.StatusNotFound, result
	}

	filename := pubkeysPath + "/" + files[0].Name()
	pubkey, err = ioutil.ReadFile(filename)
	if err != nil {
		log.Error("[ACI API] Read pubkey file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read pubkey file failed"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

	acifile := strings.Trim(acifilename, ".asc")
	imagename := strings.Trim(acifile, ".aci")

	r := new(models.Repository)
	if has, _, err := r.Has(namespace, imagename); err != nil {
		log.Error("[ACI API] Read ACI %v/%v detail error: %v", namespace, imagename, err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Read ACI detail error"})
		return http.StatusInternalServerError, result
	} else if has == false {
		log.Error("[ACI API] Not found ACI %v/%v", namespace, imagename)
		result, _ := json.Marshal(map[string]string{"message": "Not found ACI"})
		return http.StatusNotFound, result
	}

	var imagepath string
	if b := strings.Contains(acifilename, ".asc"); b == true {
		imagepath = r.Aci.SignPath
	} else {
		imagepath = r.Aci.AciPath
	}

	img, err := ioutil.ReadFile(imagepath)
	if err != nil {
		log.Error("[ACI API] Get ACI file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get ACI file failed"})
		return http.StatusInternalServerError, result
	}

	return http.StatusOK, img
}

func PostUploadHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	domains := ctx.Params(":domains")
	namespace := ctx.Params(":namespace")
	acifile := ctx.Params(":acifile")

	signfile := fmt.Sprintf("%v%v", acifile, ".asc")
	//aciname := strings.Trim(acifile, ".aci")

	//TODO: only for testing,pubkey will be read and saved via user management module
	pubkeyspath := fmt.Sprintf("%v/acis/pubkeys/%v", setting.ImagePath, namespace)
	if _, err := os.Stat(pubkeyspath); err != nil {
		if err := os.MkdirAll(pubkeyspath, os.ModePerm); err != nil {
			log.Error("[ACI API] Create pubkeys path failed: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Create pubkeys path failed"})
			return http.StatusInternalServerError, result
		}
	}

	aciid := utils.MD5(uuid.NewV4().String())
	imagepath := fmt.Sprintf("%v/acis/%v", setting.ImagePath, aciid)
	if err := os.MkdirAll(imagepath, os.ModePerm); err != nil {
		log.Error("[ACI API] Create aci path failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Create aci path failed"})
		return http.StatusInternalServerError, result
	}

	prefix := fmt.Sprintf("%v://%v/ac-push/%v/", setting.ListenMode, domains, namespace)
	endpoint := models.UploadDetails{
		ACIPushVersion: "0.0.1", //TODO:It would follow APPC spec
		Multipart:      false,
		ManifestURL:    prefix + aciid + "/manifest",
		SignatureURL:   prefix + aciid + "/signature/" + signfile,
		ACIURL:         prefix + aciid + "/aci/" + acifile,
		CompletedURL:   prefix + aciid + "/complete/" + acifile + "/" + signfile,
	}

	result, _ := json.Marshal(endpoint)
	return http.StatusOK, result
}

func PutManifestHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	aciid := ctx.Params(":aciid")

	manipath := fmt.Sprintf("%v/acis/%v/manifest", setting.ImagePath, aciid)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(manipath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Save manifest failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Save manifest failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	signfile := ctx.Params(":signfile")
	aciid := ctx.Params(":aciid")

	signpath := fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, aciid, signfile)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signpath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Save signature file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Save signature file failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	acifile := ctx.Params(":acifile")
	aciid := ctx.Params(":aciid")

	acipath := fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, aciid, acifile)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(acipath, data, 0777); err != nil {
		//Temporary directory would be deleted in PostCompleteHandler
		log.Error("[ACI API] Save aci file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Save aci file failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostCompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	aciid := ctx.Params(":aciid")

	body, _ := ctx.Req.Body().Bytes()
	if err := module.CheckClientStatus(body); err != nil {
		module.CleanCache(aciid)

		log.Error("[ACI API] Push aci failed: %v", err.Error())
		failmsg := module.FillRespMsg(false, err.Error(), "")
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	namespace := ctx.Params(":namespace")
	acifile := ctx.Params(":acifile")
	signfile := ctx.Params(":signfile")

	//TODO: only for testing,pubkey will be read and saved via user management module
	pubkeyspath := fmt.Sprintf("%v/acis/pubkeys/%v", setting.ImagePath, namespace)

	acipath := fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, aciid, acifile)
	signpath := fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, aciid, signfile)
	manipath := fmt.Sprintf("%v/acis/%v/manifest", setting.ImagePath, aciid)
	if err := module.CheckAciSignature(acipath, signpath, pubkeyspath); err != nil {
		module.CleanCache(aciid)

		log.Error("[ACI API] Aci check failed: %v", err.Error())
		failmsg := module.FillRespMsg(false, "", err.Error())
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	aciname := strings.Trim(acifile, ".aci")
	r := new(models.Repository)
	has, _, err := r.Has(namespace, aciname)
	if err != nil {
		module.CleanCache(aciid)

		log.Error("[ACI API] Get %v/%v failed: %v", namespace, aciname, err.Error())
		failmsg := module.FillRespMsg(false, "", err.Error())
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	//The old aci directory should be deleted soon
	var oldaciid = ""
	if has == true {
		oldaciid = r.Aci.AciID
	}

	if err := r.Update(namespace, aciname, aciid, manipath, signpath, acipath); err != nil {
		module.CleanCache(aciid)

		log.Error("[ACI API] Update %v/%v failed: %v", namespace, aciname, err.Error())
		failmsg := module.FillRespMsg(false, "", err.Error())
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	//Delete old aci directory after redis is updated
	if oldaciid != "" {
		module.CleanCache(oldaciid)
	}

	successmsg := module.FillRespMsg(true, "", "")
	result, _ := json.Marshal(successmsg)
	return http.StatusOK, result
}
