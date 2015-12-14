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

	files, err := ioutil.ReadDir(r.PubKeysPath)
	if err != nil {
		log.Error("[ACI API] Search pubkey file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Search pubkey file failed"})
		return http.StatusNotFound, result
	}

	// TODO: consider to deal with case that one user has mutiple pubkeys in the future
	pubkey, err := ioutil.ReadFile(files[0].Name())
	if err != nil {
		log.Error("[ACI API] Read pubkey file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get pubkey file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

	r := new(models.AciRepository)
	if err := r.GetRepository(namespace); err != nil {
		log.Error("[ACI API] Get user %v details failed: %v", namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get user details failed"})
		return http.StatusNotFound, result
	}

	imgpath := ""
	imgname := strings.Split(acifilename, ".")[0]
	for _, aci := range r.Acis {
		if aci.ImageName == imgname {
			if b := strings.Contains(acifilename, ".asc"); b == true {
				imgpath = aci.SignPath
			} else {
				imgpath = aci.AciPath
			}
			break
		}
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
