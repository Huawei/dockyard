package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

func GetUsersV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusOK, result
}

func PostUsersV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	result, _ := json.Marshal(map[string]string{})
	return http.StatusUnauthorized, result
}
