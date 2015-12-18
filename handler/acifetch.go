package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
)

func GetPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")

	r := new(models.AciRepository)
	if err := r.GetRepository(namespace); err != nil {
		log.Error("[ACI API] Get user %v details failed: %v", namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get user details failed"})
		return http.StatusNotFound, result
	}
    
    //TODO:save and update interface for pubkeys source
	r.PubKeysPath = "/home/gopath/src/github.com/containerops/dockyard/data/acipool/pzh"

	files, err := ioutil.ReadDir(r.PubKeysPath)
	if err != nil {
		log.Error("[ACI API] Search pubkey file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Search pubkey file failed"})
		return http.StatusNotFound, result
	}

    var pubkey []byte
	// TODO: considering that one user has multiple pubkeys
    for _, file := range files {
    	if b := strings.Contains(file.Name(), ".gpg"); b == true {
    		filename := r.PubKeysPath + "/" + file.Name()
			pubkey, err = ioutil.ReadFile(filename)
			if err != nil {
				log.Error("[ACI API] Read pubkey file failed: %v", err.Error())

				result, _ := json.Marshal(map[string]string{"message": "Get pubkey file failed"})
				return http.StatusNotFound, result
			} else {
				break  
			}		 		
    	}
    }
	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

	//cut .asc and .aci of acifilename
    nameTemp := strings.Trim(acifilename, ".asc")
	imgname := strings.Trim(nameTemp, ".aci")


	var err error
    aci := &models.AciDetail{}

	r := new(models.AciRepository)
	if aci, err = r.GetAciByName(namespace, imgname); err != nil {
		log.Error("[ACI API] Get aci %v details failed: %v", namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get aci details failed"})
		return http.StatusNotFound, result
	}

    var imgpath string
	if b := strings.Contains(acifilename, ".asc"); b == true {
		imgpath = aci.SignPath + "/" + acifilename
	} else {
		imgpath = aci.AciPath+ "/" + acifilename
	}
    
	//imgpath := setting.ImagePath + "/acipool/" + aciname
	img, err := ioutil.ReadFile(imgpath)
	if err != nil {
		log.Error("[ACI API] Read ACI file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get ACI file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, img
}
