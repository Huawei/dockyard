package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/macaron"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/utils"
	"net/http"
)

func GetUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	TmpPrepare(ctx)

	if username, passwd, err := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization")); err != nil {
		fmt.Printf("[REGISTRY API V1] Decode Basic Auth Error:%v", err.Error())

		result, _ := json.Marshal(map[string]string{"error": "Decode authorization failure"})
		return http.StatusUnauthorized, result
	} else {
		user := new(models.User)
		if err := user.Get(username, passwd); err != nil {
			fmt.Printf("[REGISTRY API V1] Search user error: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"error": "User authorization failure"})
			return http.StatusUnauthorized, result
		}

		fmt.Printf("[REGISTRY API V1] %v authorization successfully", username)
		/*
			memo, _ := json.Marshal(this.Ctx.Input.Header)
			if err := user.Log(models.ACTION_SIGNUP, models.LEVELINFORMATIONAL, models.TYPE_APIV1, user.UUID, memo); err != nil {
				fmt.Printf("[REGISTRY API V1] Log Erro:", err.Error())
			}
		*/

		result, _ := json.Marshal(map[string]string{"status": "User authorization successfully"})
		return http.StatusOK, result
	}
}

func PostUsersV1Handler(ctx *macaron.Context) (int, []byte) {
	TmpPrepare(ctx)

	result, _ := json.Marshal(map[string]string{"message": ""})
	return http.StatusUnauthorized, result
}
