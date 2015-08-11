package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

func GetPingV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte("")
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	return http.StatusOK, []byte("")
}
