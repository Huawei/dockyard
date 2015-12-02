package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/wrench/setting"
)


type aci struct {
	Name    string
	Details []acidetails
}

type acidetails struct {
	Version string
	OS      string
	Arch    string
	Signed  bool
	LastMod string
}

type initiateDetails struct {
	ACIPushVersion string `json:"aci_push_version"`
	Multipart      bool   `json:"multipart"`
	ManifestURL    string `json:"upload_manifest_url"`
	SignatureURL   string `json:"upload_signature_url"`
	ACIURL         string `json:"upload_aci_url"`
	CompletedURL   string `json:"completed_url"`
}

type completeMsg struct {
	Success      bool   `json:"success"`
	Reason       string `json:"reason,omitempty"`
	ServerReason string `json:"server_reason,omitempty"`
}

type upload struct {
	Started time.Time
	Image   string
	GotSig  bool
	GotACI  bool
	GotMan  bool
}

var (
	directory     string
	templatedir   string
	uploadcounter int
	newuploadLock sync.Mutex
	uploads       map[int]*upload
)

func init() {
	uploads = make(map[int]*upload)

	directory = setting.ImagePath + "/acpool/"
	templatedir = "conf"

}

func RenderListOfACIs(ctx *macaron.Context, log *logs.BeeLogger) {
	os.RemoveAll(path.Join(directory, "tmp"))
	err := os.MkdirAll(path.Join(directory, "tmp"), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	t, err := template.ParseFiles(path.Join(templatedir, "acipushtemplate.html"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	acis, err := listACIs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	err = t.Execute(ctx.Resp, struct {
		ServerName string
		ACIs       []aci
		Domain     string
	}{
		ServerName: setting.Domains,
		ACIs:       acis,
		Domain:     setting.ListenMode,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
}

func GetPubkeys(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var pubkey []byte
	var err error

	pubkeypath := setting.ImagePath + "/acpool/" + "pubkeys.gpg"
	if pubkey, err = ioutil.ReadFile(pubkeypath); err != nil {
		log.Error("[ACI API] Get pubkey file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Get pubkey file failed"})
		return http.StatusNotFound, result
	}
	return http.StatusOK, pubkey
}

func InitiateUpload(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	image := ctx.Params(":image")
	if image == "" {
		log.Error("[ACI API]Get image name failed")
		result, _ := json.Marshal(map[string]string{"message": "Get image name failed"})
		return http.StatusNotFound, result
	}

	uploadNum := strconv.Itoa(newUpload(image))

	var prefix string
	prefix = setting.ListenMode+"://" + setting.Domains + "/ac-push" 

	deets := initiateDetails{
		ACIPushVersion: "0.0.1",
		Multipart:      false,
		ManifestURL:    prefix + "/manifest/" + uploadNum,
		SignatureURL:   prefix + "/signature/" + uploadNum,
		ACIURL:         prefix + "/aci/" + uploadNum,
		CompletedURL:   prefix + "/complete/" + uploadNum,
	}

	result, _ := json.Marshal(deets)
	return http.StatusInternalServerError, result

}

func UploadManifest(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	num, err := strconv.Atoi(ctx.Params(":num"))
	if err != nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	err = gotMan(num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func ReceiveSignUpload(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	num, err := strconv.Atoi(ctx.Params(":num"))
	if err != nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	up := getUpload(num)
	if up == nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	_, err = os.Stat(up.Image)
	if err == nil {
		log.Error("[ACI API]item already uploaded")
		result, _ := json.Marshal(map[string]string{"message": "item already uploaded"})
		return http.StatusConflict, result
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	aci, err := os.OpenFile(tmpSigPath(num),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}
	defer aci.Close()

	_, err = io.Copy(aci, ctx.Req.Request.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	err = gotSig(num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func ReceiveAciUpload(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	num, err := strconv.Atoi(ctx.Params(":num"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	up := getUpload(num)
	if up == nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	_, err = os.Stat(up.Image)
	if err == nil {
		log.Error("[ACI API]item already uploaded")
		result, _ := json.Marshal(map[string]string{"message": "item already uploaded"})
		return http.StatusConflict, result
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	aci, err := os.OpenFile(tmpACIPath(num),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}
	defer aci.Close()

	_, err = io.Copy(aci, ctx.Req.Request.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	err = gotACI(num)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func tmpSigPath(num int) string {
	return path.Join(directory, "tmp", strconv.Itoa(num)+".asc")
}

func tmpACIPath(num int) string {
	return path.Join(directory, "tmp", strconv.Itoa(num))
}

func CompleteUpload(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	num, err := strconv.Atoi(ctx.Params(":num"))
	if err != nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	up := getUpload(num)
	if up == nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusNotFound, result
	}

	body, err := ioutil.ReadAll(ctx.Req.Request.Body)
	if err != nil {
		result, _ := json.Marshal(map[string]string{})
		return http.StatusInternalServerError, result
	}

	fmt.Fprintf(os.Stderr, "body: %s\n", string(body))

	msg := completeMsg{}
	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Error("[ACI API]error unmarshaling json: %v", err.Error())
		result, _ := json.Marshal("error unmarshaling json")
		return http.StatusBadRequest, result
	}

	if !msg.Success {
		err := reportFailure(num)
		if err != nil {
		    log.Error("[ACI API]client reported failure: %v", err.Error())
			status, result := msgFailure("client reported failure", msg.Reason)
			return status, result
		}
	}

	if !up.GotMan {
		err := reportFailure(num)
		if err != nil {
		    log.Error("[ACI API]manifest wasn't uploaded: %v", err.Error())
			status, result := msgFailure("manifest wasn't uploaded", msg.Reason)
			return status, result
		}
	}

	if !up.GotSig {
		err := reportFailure(num)
		if err != nil {
		    log.Error("[ACI API]signature wasn't uploaded: %v", err.Error())
			status, result := msgFailure("signature wasn't uploaded", msg.Reason)
			return status, result
		}
	}

	if !up.GotACI {
		err := reportFailure(num)
		if err != nil {
		    log.Error("[ACI API]ACI wasn't uploaded: %v", err.Error())
			status, result := msgFailure("ACI wasn't uploaded", msg.Reason)
			return status, result
		}
	}

	err = finishUpload(num)
	if err != nil {
		err := reportFailure(num)
		if err != nil {
		    log.Error("[ACI API]Internal Server Error: %v", err.Error())
			status, result := msgFailure("Internal Server Error", msg.Reason)
			return status, result
		}
	}

	succmsg := completeMsg{
		Success: true,
	}

	result, _ := json.Marshal(succmsg)
	return http.StatusInternalServerError, result
}

func reportFailure(num int) error {
	err := abortUpload(num)
	if err != nil {
		return err
	}
	return nil
}

func msgFailure(msg, clientmsg string) (int, []byte) {
	failmsg := completeMsg{
		Success:      false,
		Reason:       clientmsg,
		ServerReason: msg,
	}
	result, _ := json.Marshal(failmsg)
	return http.StatusInternalServerError, result
}

func abortUpload(num int) error {
	newuploadLock.Lock()
	delete(uploads, num)
	newuploadLock.Unlock()

	tmpaci := path.Join(directory, "tmp", strconv.Itoa(num))
	_, err := os.Stat(tmpaci)
	if err == nil {
		err = os.Remove(tmpaci)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	tmpsig := path.Join(directory, "tmp", strconv.Itoa(num)+".asc")
	_, err = os.Stat(tmpsig)
	if err == nil {
		err = os.Remove(tmpsig)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func finishUpload(num int) error {
	newuploadLock.Lock()
	up, ok := uploads[num]
	if ok {
		delete(uploads, num)
	}
	newuploadLock.Unlock()
	if !ok {
		return fmt.Errorf("no such upload: %d", num)
	}

	err := os.Rename(path.Join(directory, "tmp", strconv.Itoa(num)),
		path.Join(directory, up.Image))
	if err != nil {
		return err
	}

	err = os.Rename(path.Join(directory, "tmp", strconv.Itoa(num)+".asc"),
		path.Join(directory, up.Image+".asc"))
	if err != nil {
		return err
	}

	return nil
}

func newUpload(image string) int {
	newuploadLock.Lock()
	uploadcounter++
	uploads[uploadcounter] = &upload{
		Started: time.Now(),
		Image:   image,
	}
	newuploadLock.Unlock()
	return uploadcounter
}

func getUpload(num int) *upload {
	var up *upload
	newuploadLock.Lock()
	up, ok := uploads[num]
	newuploadLock.Unlock()
	if !ok {
		return nil
	}
	return up
}

func gotSig(num int) error {
	newuploadLock.Lock()
	_, ok := uploads[num]
	if ok {
		uploads[num].GotSig = true
	}
	newuploadLock.Unlock()
	if !ok {
		return fmt.Errorf("no such upload: %d", num)
	}
	return nil
}

func gotACI(num int) error {
	newuploadLock.Lock()
	_, ok := uploads[num]
	if ok {
		uploads[num].GotACI = true
	}
	newuploadLock.Unlock()
	if !ok {
		return fmt.Errorf("no such upload: %d", num)
	}
	return nil
}

func gotMan(num int) error {
	newuploadLock.Lock()
	_, ok := uploads[num]
	if ok {
		uploads[num].GotMan = true
	}
	newuploadLock.Unlock()
	if !ok {
		return fmt.Errorf("no such upload: %d", num)
	}
	return nil
}

func listACIs() ([]aci, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var acis []aci
	for _, file := range files {
		_, fname := path.Split(file.Name())
		tokens := strings.Split(fname, "-")
		if len(tokens) != 4 {
			continue
		}

		tokens1 := strings.Split(tokens[3], ".")
		if len(tokens1) != 2 {
			continue
		}

		if tokens1[1] != "aci" {
			continue
		}

		var signed bool

		_, err := os.Stat(path.Join(directory, fname+".asc"))
		if err == nil {
			signed = true
		} else if os.IsNotExist(err) {
			signed = false
		} else {
			return nil, err
		}

		details := acidetails{
			Version: tokens[1],
			OS:      tokens[2],
			Arch:    tokens1[0],
			Signed:  signed,
			LastMod: file.ModTime().Format("Mon Jan 2 15:04:05 -0700 MST 2006"),
		}

		// If the last ACI added to the list has the same name
		if len(acis) > 0 && acis[len(acis)-1].Name == tokens[0] {
			acis[len(acis)-1].Details = append(acis[len(acis)-1].Details,
				details)
		} else {
			acis = append(acis, aci{
				Name:    tokens[0],
				Details: []acidetails{details},
			})
		}
	}

	return acis, nil
}
