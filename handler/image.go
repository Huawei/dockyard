package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Unknwon/macaron"
)

func GetImageAncestryV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetImageLayerV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func PutImageJSONV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func PutImageLayerv1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func PutImageChecksumV1Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}
