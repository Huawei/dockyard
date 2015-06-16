package handler

import (
	"encoding/json"
	"github.com/Unknwon/macaron"
	"net/http"
)

func GetPingV1Handler(ctx *macaron.Context) (int, []byte) {

	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	//TBD:not support V2 ping now,return 404 for test
	return 404, result
}
