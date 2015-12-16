package handler

import (
	"encoding/json"
	"fmt"
//	"io"
	"io/ioutil"
	"net/http"
	"os"
//	"path"
//	"strconv"
	"strings"
//	"sync"
//	"time"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/wrench/setting"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/utils"
)

var acideets *models.AciDetail


func PutPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    //TODO:check user`s uploaded pubkey is exist or not, save and append to files

	result, _ := json.Marshal(map[string]string{})
	return http.StatusCreated, result
}

func GetUploadDetailsHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {     
	namespace := ctx.Params(":namespace")
	servername := ctx.Params(":servername")
	acifilename := ctx.Params(":acifile")
	imgname := strings.Trim(acifilename, ".aci")

	var err error

    r := new(models.AciRepository)
	if acideets, err = r.GetAciByName(namespace, imgname); err == nil {
		log.Error("[ACI API] aci %v is existed in %v repository", acifilename, namespace)

		result, _ := json.Marshal(map[string]string{"message": "aci is existed"})
		return http.StatusInternalServerError, result
	}

    prefix := setting.ListenMode + "://" + servername
     
	deets := models.UploadDetails{
		ACIPushVersion: "0.0.1",
		Multipart:      false,
		ManifestURL:    prefix + "/manifest/" + acifilename,
		SignatureURL:   prefix + "/sign/" + acifilename,
		ACIURL:         prefix + "/aci/" + acifilename,
		CompletedURL:   prefix + "/complete/",
	}

	result, _ := json.Marshal(deets)
	return http.StatusOK, result
}

func PutManifestHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
 //   namespace := ctx.Params(":namespace")
 //	  acifilename := ctx.Params(":acifile")
 //	  imgname := strings.Trim(acifilename, ".aci")

    //TODO:save manifest to repository
    
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) { 
    namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")
	imgname := strings.Trim(acifilename, ".aci")
	signfilename := imgname + ".aci.asc"

	signPathTmp := fmt.Sprintf("%v/acpool/%v/%v", setting.ImagePath, namespace, imgname)
	signfileTmp := fmt.Sprintf("%v/acpool/%v/%v/%v", setting.ImagePath, namespace, imgname, signfilename)

	if !utils.IsDirExist(signPathTmp) {
		os.MkdirAll(signPathTmp, os.ModePerm)
	}

	if _, err := os.Stat(signfileTmp); err == nil {
		os.Remove(signfileTmp)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signfileTmp, data, 0777); err != nil {
		log.Error("[ACI API] Save signaturefile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save signaturefile failed"})
		return http.StatusBadRequest, result
	}	

	acideets.SignPath = signPathTmp

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")
	imgname := strings.Trim(acifilename, ".aci")

	aciPathTmp := fmt.Sprintf("%v/acpool/%v/%v", setting.ImagePath, namespace, imgname)
	acifileTmp := fmt.Sprintf("%v/acpool/%v/%v/%v", setting.ImagePath, namespace, imgname, acifilename)

	if !utils.IsDirExist(aciPathTmp) {
		os.MkdirAll(aciPathTmp, os.ModePerm)
	}

	if _, err := os.Stat(acifileTmp); err == nil {
		os.Remove(acifileTmp)
	}

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(acifileTmp, data, 0777); err != nil {
		log.Error("[ACI API] Save signaturefile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save signaturefile failed"})
		return http.StatusBadRequest, result
	}	

	acideets.AciPath = aciPathTmp

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func CompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    namespace := ctx.Params(":namespace")
    //TODO: image verification here

    r := new(models.AciRepository)
    if err := r.PutAciByName(namespace, acideets.ImageName, acideets); err != nil {
        log.Error("[ACI API] Save aci %v details to %v repository failed: %v", acideets.ImageName, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get aci details failed"})
		return http.StatusNotFound, result
    }
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}



