package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego/logs"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func PostUploadHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	domains := ctx.Params(":domains")
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

	signfilename := fmt.Sprintf("%v%v", acifilename, ".asc")
	imgname := strings.Trim(acifilename, ".aci")

	aciPathTmp := fmt.Sprintf("%v/acipool/%v/tmp", setting.ImagePath, namespace)
	aciPath := fmt.Sprintf("%v/acipool/%v/%v", setting.ImagePath, namespace, imgname)
	pubkeysPath := fmt.Sprintf("%v/acipool/%v/pubkeys", setting.ImagePath, namespace)

	//handle tmp dir
	if err := os.RemoveAll(aciPathTmp); err != nil {
		log.Error("[ACI API] Remove aciPathTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Remove aciPathTmp failed"})
		return http.StatusNotFound, result
	}

	if err := os.MkdirAll(aciPathTmp, os.ModePerm); err != nil {
		log.Error("[ACI API] Make aciPathTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Make aciPathTmp failed"})
		return http.StatusInternalServerError, result
	}

	//handle aci dir
	if !utils.IsDirExist(aciPath) {
		if err := os.MkdirAll(aciPath, os.ModePerm); err != nil {
			log.Error("[ACI API] Make aciPath failed: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Make aciPath failed"})
			return http.StatusNotFound, result
		}
		//acipath will be deleted in the end by this mark, default true
		models.AcipathExist = false
	}

	//handle pubkeys dir
	if !utils.IsDirExist(pubkeysPath) {
		if err := os.MkdirAll(pubkeysPath, os.ModePerm); err != nil {
			log.Error("[ACI API] Make pubkeysPath failed: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Make pubkeysPath failed"})
			return http.StatusBadRequest, result
		}
	}

	prefix := fmt.Sprintf("%v://%v/ac-push/%v", setting.ListenMode, domains, namespace)

	endpoint := models.UploadDetails{
		ACIPushVersion: "0.0.1", //TODO:It would follow APPC spec
		Multipart:      false,
		ManifestURL:    prefix + "/manifest/" + imgname,
		SignatureURL:   prefix + "/signature/" + signfilename,
		ACIURL:         prefix + "/aci/" + acifilename,
		CompletedURL:   prefix + "/complete/" + imgname,
	}

	result, _ := json.Marshal(endpoint)
	return http.StatusOK, result
}

func PutManifestHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	aciPathTmp := fmt.Sprintf("%v/acipool/%v/tmp", setting.ImagePath, namespace)
	maniFullnameTmp := fmt.Sprintf("%v/manifest", aciPathTmp)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(maniFullnameTmp, data, 0777); err != nil {
		log.Error("[ACI API] Save manifileTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save manifileTmp failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	signfilename := ctx.Params(":acifile")

	aciPathTmp := fmt.Sprintf("%v/acipool/%v/tmp", setting.ImagePath, namespace)
	signFullnameTmp := fmt.Sprintf("%v/%v", aciPathTmp, signfilename)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signFullnameTmp, data, 0777); err != nil {
		log.Error("[ACI API] Save signaturefileTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save signaturefileTmp failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

	aciPathTmp := fmt.Sprintf("%v/acipool/%v/tmp", setting.ImagePath, namespace)
	aciFullnameTmp := fmt.Sprintf("%v/%v", aciPathTmp, acifilename)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(aciFullnameTmp, data, 0777); err != nil {
		log.Error("[ACI API] Save acifileTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save acifileTmp failed"})
		return http.StatusInternalServerError, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostCompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	imgname := ctx.Params(":acifile")

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Get request body failed"})
		return http.StatusInternalServerError, result
	}

	clientmsg := new(models.CompleteMsg)
	if err := json.Unmarshal(body, &clientmsg); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Unmarshal failed"})
		return http.StatusInternalServerError, result
	}

	//aci image check
	httpstatus, checkresult, err := ImageCheck(namespace, imgname, log)
	if err != nil {
		log.Error("[ACI API] Aci image check failed: %v", err.Error())

		failmsg := models.CompleteMsg{
			Success:      false,
			Reason:       "ACI is invalid",
			ServerReason: string(checkresult),
		}
		result, _ := json.Marshal(failmsg)
		return httpstatus, result
	}

	succmsg := models.CompleteMsg{
		Success: true,
	}
	result, _ := json.Marshal(succmsg)
	return http.StatusOK, result
}

func ImageCheck(namespace string, imgname string, log *logs.BeeLogger) (int, []byte, error) {
	aciPathTmp := fmt.Sprintf("%v/acipool/%v/tmp", setting.ImagePath, namespace)
	aciPath := fmt.Sprintf("%v/acipool/%v/%v", setting.ImagePath, namespace, imgname)

	signFullnameTmp := fmt.Sprintf("%v/%v%v", aciPathTmp, imgname, ".aci.asc")
	aciFullnameTmp := fmt.Sprintf("%v/%v%v", aciPathTmp, imgname, ".aci")

	signFullname := fmt.Sprintf("%v/%v%v", aciPath, imgname, ".aci.asc")
	aciFullname := fmt.Sprintf("%v/%v%v", aciPath, imgname, ".aci")
	acifromPushname := fmt.Sprintf("%v/%v/%v", setting.Domains, namespace, strings.Split(imgname, "-")[0])

	keyspath := fmt.Sprintf("%v/acipool/%v/pubkeys", setting.ImagePath, namespace)

	//image verification
	if err := ImageVerification(signFullnameTmp, aciFullnameTmp, acifromPushname, keyspath); err != nil {
		//remove aci tmp and new created dir
		if err := RemoveDir(aciPathTmp, aciPath); err != nil {
			log.Error("[ACI API] Remove Aci dir failed: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Remove Aci dir failed"})
			return http.StatusBadRequest, result, err
		}

		log.Error("[ACI API] Aci Verification failed : %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Aci Verification failed"})
		return http.StatusBadRequest, result, err
	}

	//save to db
	r := new(models.AciRepository)
	if err := r.PutAciByName(namespace, imgname, signFullname, aciFullname, keyspath); err != nil {
		log.Error("[ACI API] Save aci %v details to %v repository failed: %v", imgname, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save aci details failed"})
		return http.StatusNotFound, result, err
	}

	//copy aci files
	if err := MoveAcifiles(signFullname, aciFullname, signFullnameTmp, aciFullnameTmp); err != nil {
		log.Error("[ACI API] Move Acifiles failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Move Acifiles failed"})
		return http.StatusBadRequest, result, err
	}

	//remove aci tmp dir
	if err := os.RemoveAll(aciPathTmp); err != nil {
		log.Error("[ACI API] Remove aciPathTmp failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Remove aciPathTmp failed"})
		return http.StatusBadRequest, result, err
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result, nil
}

func ImageVerification(signFullnameTmp string, aciFullnameTmp string, acifromPushname string, keyspath string) error {
	signfileTmp, err := os.Open(signFullnameTmp)
	if err != nil {
		return fmt.Errorf("opening signfileTmp file failed: %v", err.Error())
	}
	defer signfileTmp.Close()

	acifileTmp, err := os.Open(aciFullnameTmp)
	if err != nil {
		return fmt.Errorf("opening acifileTmp file failed: %v", err.Error())
	}
	defer acifileTmp.Close()

	//load keyring
	files, err := ioutil.ReadDir(keyspath)
	if err != nil {
		return fmt.Errorf("Search pubkey file failed: %v", err.Error())
	}

	var keyring openpgp.EntityList
	trustedKeys := make(map[string]*openpgp.Entity)

	for _, file := range files {
		keypath := fmt.Sprintf("%v/%v", keyspath, file.Name())
		pubKeyfile, err := os.Open(keypath)
		if err != nil {
			return err
		}
		defer pubKeyfile.Close()
		keyList, err := openpgp.ReadArmoredKeyRing(pubKeyfile)
		if err != nil {
			return err
		}
		if len(keyList) < 1 {
			return fmt.Errorf("missing opengpg entity")
		}

		fingerprint := fmt.Sprintf("%x", keyList[0].PrimaryKey.Fingerprint)
		if fingerprint != file.Name() {
			return fmt.Errorf("fingerprint mismatch: %v:%v", file.Name(), fingerprint)
		}

		trustedKeys[fingerprint] = keyList[0]
	}

	for _, v := range trustedKeys {
		keyring = append(keyring, v)
	}

	//check keyring asc aci
	if _, err := signfileTmp.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking ACI file: %v", err)
	}
	if _, err := acifileTmp.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking signature file: %v", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, acifileTmp, signfileTmp)
	if err == io.EOF {
		if _, err := signfileTmp.Seek(0, 0); err != nil {
			return fmt.Errorf("error seeking ACI file: %v", err)
		}
		if _, err := acifileTmp.Seek(0, 0); err != nil {
			return fmt.Errorf("error seeking signature file: %v", err)
		}

		_, err = openpgp.CheckDetachedSignature(keyring, acifileTmp, signfileTmp)
	}
	if err == io.EOF {
		return fmt.Errorf("no valid signatures found in signature file")
	}
	return nil
}

func MoveAcifiles(signFullname string, aciFullname string, signFullnameTmp string, aciFullnameTmp string) error {
	//orverride signfile
	signfile, err := os.OpenFile(signFullname, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening signfile failed: %v", err.Error())
	}

	signfileTmp, err := os.Open(signFullnameTmp)
	if err != nil {
		return fmt.Errorf("opening signfileTmp failed: %v", err.Error())
	}
	defer signfileTmp.Close()

	if _, err := io.Copy(signfile, signfileTmp); err != nil {
		return fmt.Errorf("override signfile failed: %v", err.Error())
	}

	//orverride acifile
	acifile, err := os.OpenFile(aciFullname, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("opening signfile failed: %v", err.Error())
	}

	acifileTmp, err := os.Open(aciFullnameTmp)
	if err != nil {
		return fmt.Errorf("opening acifileTmp failed: %v", err.Error())
	}
	defer acifileTmp.Close()

	if _, err := io.Copy(acifile, acifileTmp); err != nil {
		return fmt.Errorf("override acifile failed: %v", err.Error())
	}
	return nil
}

func RemoveDir(aciPathTmp string, aciPath string) error {
	//remove aci tmp dir
	if err := os.RemoveAll(aciPathTmp); err != nil {
		return fmt.Errorf("Remove aciPathTmp failed: %v", err.Error())
	}

	//remove new created aci dir
	if !models.AcipathExist {
		//refresh mark to default
		models.AcipathExist = true
		if err := os.RemoveAll(aciPath); err != nil {
			return fmt.Errorf("Remove new created aciPath failed: %v", err.Error())
		}
	}
	return nil
}
