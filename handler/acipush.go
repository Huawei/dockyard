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

	"github.com/containerops/wrench/setting"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/utils"
)

func PutPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    //TODO:check user`s uploaded pubkey is existed or not, save file and append to keyring

	result, _ := json.Marshal(map[string]string{})
	return http.StatusCreated, result
}

func GetUploadEndPointHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {     
	servername := ctx.Params(":servername")
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")
	imgname := strings.Trim(acifilename, ".aci")

    //TODO:check aci is existed or not and handle that, need acpush clent`s cooperation

    prefix := setting.ListenMode + "://" + servername + "/ac-push/" + namespace
     
	endpoint := models.UploadDetails{
		ACIPushVersion: setting.AcipushVersion,  
		Multipart:      false,
		ManifestURL:    prefix + "/manifest/" + imgname,
		SignatureURL:   prefix + "/signature/" + imgname,
		ACIURL:         prefix + "/aci/" + imgname,
		CompletedURL:   prefix + "/complete/" + imgname,
	}

	result, _ := json.Marshal(endpoint)
	return http.StatusOK, result
}

func PutManifestHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    namespace := ctx.Params(":namespace")
 	imgname := ctx.Params(":imgname")

    data, _ := ctx.Req.Body().Bytes()
    manifest := string(data)

	r := new(models.AciRepository)
    if err := r.PutManifest(namespace, imgname, manifest); err != nil {
        log.Error("[ACI API] Save aci %v details to %v repository failed: %v", imgname, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save aci details failed"})
		return http.StatusNotFound, result
    }

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
    namespace := ctx.Params(":namespace")
	imgname := ctx.Params(":imgname")
	signfilename := imgname + ".aci.asc"

	signpath := fmt.Sprintf("%v/acipool/%v/%v", setting.ImagePath, namespace, imgname)
	signfile := fmt.Sprintf("%v/acipool/%v/%v/%v", setting.ImagePath, namespace, imgname, signfilename)

	if !utils.IsDirExist(signpath) {
		os.MkdirAll(signpath, os.ModePerm)
	}

	if _, err := os.Stat(signfile); err == nil {
		os.Remove(signfile)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signfile, data, 0777); err != nil {
		log.Error("[ACI API] Save signaturefile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save signaturefile failed"})
		return http.StatusBadRequest, result
	}	

	r := new(models.AciRepository)
    if err := r.PutSignpath(namespace, imgname, signpath); err != nil {
        log.Error("[ACI API] Save aci %v details to %v repository failed: %v", imgname, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save aci details failed"})
		return http.StatusNotFound, result
    }

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    namespace := ctx.Params(":namespace")
	imgname := ctx.Params(":imgname")
	acifilename := imgname + ".aci"

	acipath := fmt.Sprintf("%v/acipool/%v/%v", setting.ImagePath, namespace, imgname)
	acifile := fmt.Sprintf("%v/acipool/%v/%v/%v", setting.ImagePath, namespace, imgname, acifilename)

	if !utils.IsDirExist(acipath) {
		os.MkdirAll(acipath, os.ModePerm)
	}

	if _, err := os.Stat(acifile); err == nil {
		os.Remove(acifile)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(acifile, data, 0777); err != nil {
		log.Error("[ACI API] Save acifile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save acifile failed"})
		return http.StatusBadRequest, result
	}	

	r := new(models.AciRepository)
    if err := r.PutAcipath(namespace, imgname, acipath); err != nil {
        log.Error("[ACI API] Save aci %v details to %v repository failed: %v", imgname, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save aci details failed"})
		return http.StatusNotFound, result
    }

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func CompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {   
    namespace := ctx.Params(":namespace")
	imgname := ctx.Params(":imgname")

	var err error
    aci := &models.AciDetail{}
    
	r := new(models.AciRepository)
	if aci, err = r.GetAciByName(namespace, imgname); err != nil {
		log.Error("[ACI API] Get aci %v details failed: %v", namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get aci details failed"})
		return http.StatusNotFound, result
	}

	//TODO: image verification here

    //TODO: delete all uploaded files and redis records if verification fail

	if aci.UpMan != true {
        failmsg := models.CompleteMsg{
			Success:      false,
			Reason:       "",
			ServerReason: "manifest wasn't uploaded",
		}
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	if aci.UpSig != true {
        failmsg := models.CompleteMsg{
			Success:      false,
			Reason:       "",
			ServerReason: "signaturen wasn't uploaded",
		}
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}

	if aci.UpAci != true {
        failmsg := models.CompleteMsg{
			Success:      false,
			Reason:       "",
			ServerReason: "aci wasn't uploaded",
		}
		result, _ := json.Marshal(failmsg)
		return http.StatusInternalServerError, result
	}
     
	succmsg := models.CompleteMsg{
		Success: true,
	}
	result, _ := json.Marshal(succmsg)
	return http.StatusOK, result
}



