package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/wrench/setting"
)

/* TBD:
current implementation as blow just be added for testing ACI fetch,
they would be updated after ACI ac-push finished
*/

func GetPubkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var pubkey []byte
	var err error

	pubkeypath := setting.ImagePath + "/acpool/" + "pubkeys.gpg"
	if pubkey, err = ioutil.ReadFile(pubkeypath); err != nil {
		// TBD: consider to fetch pubkey from other storage medium

		log.Error("[ACI API] Get pubkey file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Get pubkey file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	var img []byte
	var err error

	name := ctx.Params(":acname")

	//support to fetch images from location storage, it will be supported to fetch from cloud etc.
	imgpath := setting.ImagePath + "/acpool/" + name
	if img, err = ioutil.ReadFile(imgpath); err != nil {
		// TBD: consider to fetch image from other storage medium

		log.Error("[ACI API] Get ACI file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Get ACI file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, img

}
