package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"

	"github.com/containerops/dockyard/setting"
)

func TmpPrepare(ctx *macaron.Context) {
	// TBD:just set temporary value,these will be updated after configure management be implemented
	ctx.Resp.Header().Set("Content-Type", "application/json;charset=UTF-8")
	ctx.Resp.Header().Set("X-Docker-Registry-Standalone", "True")       //Standalone
	ctx.Resp.Header().Set("X-Docker-Registry-Version", setting.Version) //Version
	ctx.Resp.Header().Set("X-Docker-Registry-Config", "dev")            //Config
	ctx.Resp.Header().Set("X-Docker-Encrypt", "false")                  //Encrypt
}

func GetPingV1Handler(ctx *macaron.Context) (int, []byte) {
	TmpPrepare(ctx)

	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return 404, result
}
