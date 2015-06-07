package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Unknwon/macaron"

	crew "github.com/containerops/crew/modules"
	"github.com/containerops/wrench/utils"
)

func GetUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	TmpPrepare(ctx)

	if username, passwd, err := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization")); err != nil {
		fmt.Printf("[DOCKER REGISTRY API V1] Decode Basic Auth Error:%v", err.Error())

		result, _ := json.Marshal(map[string]string{"error": "Decode authorization failure"})
		return http.StatusUnauthorized, result
	} else {
		if _, err := crew.GetUser(username, passwd); err != nil {
			fmt.Printf("[DOCKER REGISTRY API V1] Search user error: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"error": "User authorization failure"})
			return http.StatusUnauthorized, result
		}

		fmt.Printf("[DOCKER REGISTRY API V1] %v authorization successfully", username)

		result, _ := json.Marshal(map[string]string{"status": "User authorization successfully"})
		return http.StatusOK, result
	}
}

func PostUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	TmpPrepare(ctx)

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusUnauthorized, result
}
