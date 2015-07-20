package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Unknwon/macaron"
)

func GetPingV1Handler(ctx *macaron.Context) (int, []byte) {

	//TBD: the head value will be got from config
	ctx.Resp.Header().Set("X-Docker-Registry-Config", "dev")
	ctx.Resp.Header().Set("X-Docker-Registry-Standalone", "True")

	result, _ := json.Marshal(map[string]string{"message": "Get V1 ping success"})
	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {

	if flag := strings.Contains(ctx.Req.RequestURI, "v1"); flag == true {
		result, _ := json.Marshal(map[string]string{"message": "Get V1 request"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")

	return http.StatusOK, []byte("{}")
}
