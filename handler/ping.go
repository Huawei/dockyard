package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/macaron"
	crew "github.com/containerops/crew/modules"
	"github.com/containerops/wrench/utils"
	"net/http"
)

func GetPingV1Handler(ctx *macaron.Context) (int, []byte) {

	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetPingV2Handler(ctx *macaron.Context) (int, []byte) {
	authinfo := ctx.Req.Header.Get("Authorization")
	if len(authinfo) == 0 {
		result, _ := json.Marshal(map[string]string{"message": "Authorization missing"})
		return http.StatusUnauthorized, result

	}

	username, passwd, err := utils.DecodeBasicAuth(authinfo)
	if err != nil {
		fmt.Errorf("[REGISTRY API V2] Decode basic auth error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Verify authorization failure"})
		return http.StatusUnauthorized, result
	}

	if _, err := crew.GetUser(username, passwd); err != nil {
		fmt.Errorf("[REGISTRY API V2] Search user error: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get user error"})
		return http.StatusUnauthorized, result
	}

	fmt.Println("[REGISTRY API V2]", username, "authorization successfully")

	result, _ := json.Marshal(map[string]string{"status": "User authorization successfully"})
	return http.StatusOK, result
}
