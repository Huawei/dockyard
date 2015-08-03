package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"

	crew "github.com/containerops/crew/modules"
	"github.com/containerops/wrench/utils"
)

func GetUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	username, password, err := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	if err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Get V1 users,parse authorization failure"})
		return http.StatusUnauthorized, result
	}

	if _, err := crew.GetUser(username, password); err != nil {
		result, _ := json.Marshal(map[string]string{"message": "Get V1 users,unauthorization"})
		return http.StatusUnauthorized, result
	}

	result, _ := json.Marshal(map[string]string{"message": "Get V1 users,successfully"})
	return http.StatusOK, result
}

func PostUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": "Post V1 users,unauthorization"})
	return http.StatusUnauthorized, result
}
