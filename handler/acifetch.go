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
The current contents as blow just be added for testing ACI fetch,
they will be updated after ACI push finished
*/

// TDB: GetPukkeysHandler is not supported now, it will be updated soon
func GetPukkeysHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	pubkeypath := setting.ImagePath + "/acitest/" + "aci-pubkeys.gpg"
	pubkey, err := ioutil.ReadFile(pubkeypath)
	if err != nil {
		log.Error("[ACI API] Get pubkey file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Get pubkey file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, pubkey
}

func GetACIHandler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	name := ctx.Params(":aciname")
	imgpath := setting.ImagePath + "/acitest/" + name
	img, err := ioutil.ReadFile(imgpath)
	if err != nil {
		log.Error("[ACI API] Get ACI file failed: %v", err.Error())
		result, _ := json.Marshal(map[string]string{"message": "Get ACI file failed"})
		return http.StatusNotFound, result
	}

	return http.StatusOK, img

}
