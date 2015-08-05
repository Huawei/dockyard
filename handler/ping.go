package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"
)

func GetPingV1Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": "Get V1 ping success"})
	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	/*
		if ctx.Req.Header.Get("Authorization") == "" {
			return http.StatusUnauthorized, []byte("")
		}
	*/
	return http.StatusOK, []byte("")
}
