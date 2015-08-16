package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

func GetPingV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})

	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")

	result, _ := json.Marshal(map[string]string{})

	return http.StatusOK, result
}
