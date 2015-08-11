package handler

import (
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

func GetUsersV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte("")
}

func PostUsersV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusUnauthorized, []byte("")
}
