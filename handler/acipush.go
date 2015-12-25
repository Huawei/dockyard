package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"path"
	"path/filepath"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"
	"gopkg.in/golang.org/x/crypto/openpgp"

	"github.com/containerops/wrench/setting"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/utils"
)

func PutPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
    //TODO:load all pubkeys by web interface

    //TODO:check user`s uploaded pubkey is existed or not, save file and append to keyring

	result, _ := json.Marshal(map[string]string{})
	return http.StatusCreated, result
}

func GetUploadEndPointHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {     
	servername := ctx.Params(":servername")
	namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")
	signfilename := acifilename + ".asc"
	imgname := strings.Trim(acifilename, ".aci")

    acitmpdir := path.Join(setting.ImagePath, "acipool", namespace, "tmp")

	os.RemoveAll(acitmpdir)
	if err := os.MkdirAll(acitmpdir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

    acidir := path.Join(setting.ImagePath, "acipool", namespace, imgname)
	if !utils.IsDirExist(acidir) {
	    if err := os.MkdirAll(acidir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			result, _ := json.Marshal(map[string]string{})
			return http.StatusInternalServerError, result
		}
	}

    prefix := setting.ListenMode + "://" + servername + "/ac-push/" + namespace
     
	endpoint := models.UploadDetails{
		ACIPushVersion: setting.AcipushVersion,  
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

    acitmppath := path.Join(setting.ImagePath, "acipool", namespace, "tmp")
    manifile := path.Join(acitmppath, "manifest")

   	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(manifile, data, 0777); err != nil {
		log.Error("[ACI API] Save manifile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save manifile failed"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutSignHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
    namespace := ctx.Params(":namespace")
	signfilename := ctx.Params(":acifile")

    acitmppath := path.Join(setting.ImagePath, "acipool", namespace, "tmp")
	signfile := path.Join(acitmppath, signfilename)

	data, _ := ctx.Req.Body().Bytes()
	if err := ioutil.WriteFile(signfile, data, 0777); err != nil {
		log.Error("[ACI API] Save signaturefile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save signaturefile failed"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PutAciHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {    
    namespace := ctx.Params(":namespace")
	acifilename := ctx.Params(":acifile")

    acitmppath := path.Join(setting.ImagePath, "acipool", namespace, "tmp")
	acifile := path.Join(acitmppath, acifilename)

	data, _ := ctx.Req.Body().Bytes()   
	if err := ioutil.WriteFile(acifile, data, 0777); err != nil {
		log.Error("[ACI API] Save acifile failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save acifile failed"})
		return http.StatusBadRequest, result
	}

	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func CompleteHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {   
    namespace := ctx.Params(":namespace")
	imgname := ctx.Params(":acifile")

    body, err := ctx.Req.Body().Bytes()
    if err != nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	msg := models.CompleteMsg{}
	if err := json.Unmarshal(body, &msg); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Unmarshal failed"})
		return http.StatusBadRequest, result
	}

	fmt.Fprintf(os.Stderr, "body: %s\n", string(body))
    
    var result []byte

	httpstatus, checkresult, err := UploadCheck(namespace, imgname, log)
	if err != nil {
  		log.Error("[ACI API] UploadCheck failed: %v", err.Error())

        result, _ = ReturnFail(msg.Reason, string(checkresult), body)
	} else {
        result, _ = ReturnSuccuss()   
	} 

	return httpstatus, result
}

func ReturnFail(Reason string, checkresult string, body []byte) ([]byte, error) {  
    failmsg := models.CompleteMsg{
		Success:      false,
		Reason:       Reason,
		ServerReason: checkresult,
	}
	result, _ := json.Marshal(failmsg)
	return result, nil
}

func ReturnSuccuss() ([]byte, error) {  
	succmsg := models.CompleteMsg{
		Success: true,
	}
	result, _ := json.Marshal(succmsg)
	return result, nil
}

func UploadCheck(namespace string, imgname string, log *logs.BeeLogger) (int, []byte, error) {
	acidir   := path.Join(setting.ImagePath, "acipool", namespace, imgname)
	signpath := path.Join(acidir, imgname) + ".aci.asc"
	acipath  := path.Join(acidir, imgname) + ".aci"	

    acitmpdir := path.Join(setting.ImagePath, "acipool", namespace, "tmp")
    manitmppath := path.Join(acitmpdir, "manifest")
	signtmppath := path.Join(acitmpdir, imgname) + ".aci.asc"
	acitmppath  := path.Join(acitmpdir, imgname) + ".aci"

    manifile, err := ioutil.ReadFile(manitmppath)
	if err != nil {
		log.Error("[ACI API] opening manifile file failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Read manifile failed"})
		return http.StatusNotFound, result, err
	}

    signtmpfile, err := os.Open(signtmppath)
	if err != nil {
		log.Error("[ACI API] Read signtmpfile failed: %v", err.Error())
		
		result, _ := json.Marshal(map[string]string{"message": "Read signtmpfile failed"})
		return http.StatusNotFound, result, err
	}

    acitmpfile, err := os.Open(acitmppath)
	if err != nil {
		log.Error("[ACI API] Read acitmpfile failed: %v", err.Error())
		
		result, _ := json.Marshal(map[string]string{"message": "Read acitmpfile failed"})
		return http.StatusNotFound, result, err
	}

	keyspath := "/home/gopath/src/github.com/containerops/dockyard/data/acipool/pzh/pubkeys/pubkeys.gpg"

    if _, err := AciVerification(keyspath, manifile, signtmpfile, acitmpfile); err != nil {
	    log.Error("[ACI API] Aci Verification failed: %v", err.Error())

		if err := os.RemoveAll(acitmpdir); err != nil {
			log.Error("[ACI API] Remove acitmpdir failed: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Remove acitmpdir failed"})
			return http.StatusBadRequest, result, err
		} 	

	    result, _ := json.Marshal(map[string]string{"message": "Aci Verification failed"})
		return http.StatusBadRequest, result, err		
    } 

	//save to redis
	r := new(models.AciRepository)
    if err := r.PutAciByName(namespace, imgname, string(manifile), signpath, acipath); err != nil {
        log.Error("[ACI API] Save aci %v details to %v repository failed: %v", imgname, namespace, err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Save aci details failed"})
		return http.StatusNotFound, result, err
    }

	//delete acidir and manifest and then rename acitmpdir to right dir name
    if err := os.RemoveAll(acidir); err != nil {
		log.Error("[ACI API] Remove acidir failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Remove acidir failed"})
		return http.StatusBadRequest, result, err
	}  

    if err := os.Remove(manitmppath); err != nil {
		log.Error("[ACI API] Remove manitmppath failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Remove manitmppath failed"})
		return http.StatusBadRequest, result, err
	}  	

	if err := os.Rename(acitmpdir, acidir); err != nil {
		log.Error("[ACI API] Read manifest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Remove acidir failed"})
		return http.StatusBadRequest, result, err
	}

 	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result, nil
}  

func AciVerification(keypath string, manifile []byte, signtmpfile *os.File, acitmpfile *os.File) (*openpgp.Entity, error) {
    //check validity of manifest
	manifest := &models.ImageManifest{}
	err := json.Unmarshal(manifile, manifest)
	if err != nil {
		return nil, err
	}

	version := models.Version{}

	if manifest.ACKind != "ImageManifestKind" {
		return nil, fmt.Errorf("missing or bad ACKind (must be %v)", "ImageManifestKind")
	}
	if manifest.ACVersion == version {
		return nil, fmt.Errorf("acVersion must be set")
	}
	if string(manifest.Name) == "" {
		return nil, fmt.Errorf("name must be set")
	}

    appName := "containerops.me/pzh/etcd"

	if appName != "" && string(manifest.Name) != appName {
		return nil, fmt.Errorf("error when reading the app name: %q expected but %q found",
				appName, string(manifest.Name))
	}

	if _, err := signtmpfile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking ACI file: %v", err)
	}
	if _, err := acitmpfile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking signature file: %v", err)
	}

    //load keyring
    var keyring openpgp.EntityList
	trustedKeys := make(map[string]*openpgp.Entity)

	trustedKey, err := os.Open(keypath)
	if err != nil {
		return nil, err
	}
	defer trustedKey.Close()
	entityList, err := openpgp.ReadArmoredKeyRing(trustedKey)
	if err != nil {
		return nil, err
	}
	if len(entityList) < 1 {
		return nil, fmt.Errorf("missing opengpg entity")
	}
	fingerprint := fingerprintToFilename(entityList[0].PrimaryKey.Fingerprint)
	keyFile := filepath.Base(trustedKey.Name())
	if fingerprint != keyFile {
		return nil, fmt.Errorf("fingerprint mismatch: %q:%q", keyFile, fingerprint)
	}

	trustedKeys[fingerprintToFilename(entityList[0].PrimaryKey.Fingerprint)] = entityList[0]

	for _, v := range trustedKeys {
	    fmt.Printf("######### loadKeyring v:%v ######### \r\n", v)
		keyring = append(keyring, v)
	}

    //check keyring asc aci
	entity, err := openpgp.CheckArmoredDetachedSignature(keyring, signtmpfile, acitmpfile)
	if err == io.EOF {
		// When the signature is binary instead of armored, the error is io.EOF.
		// Let's try with binary signatures as well
		if _, err := signtmpfile.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("error seeking ACI file: %v", err)
		}
		if _, err := acitmpfile.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("error seeking signature file: %v", err)
		}
		entity, err = openpgp.CheckDetachedSignature(keyring, signtmpfile, acitmpfile)
	}
	if err == io.EOF {
		// otherwise, the client failure is just "EOF", which is not helpful
		return nil, fmt.Errorf("keystore: no valid signatures found in signature file")
	}
    return entity, err
}

func fingerprintToFilename(fp [20]byte) string {
	return fmt.Sprintf("%x", fp)
}
