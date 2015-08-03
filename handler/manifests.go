package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Unknwon/macaron"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/setting"
)

func manifestsConvertV1(data []byte) error {

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	tag := manifest["tag"]
	namespace, repository := strings.Split(manifest["name"].(string), "/")[0], strings.Split(manifest["name"].(string), "/")[1]

	for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
		v := manifest["history"].([]interface{})[k]
		compatibility := v.(map[string]interface{})["v1Compatibility"].(string)

		var image map[string]interface{}
		if err := json.Unmarshal([]byte(compatibility), &image); err != nil {
			return err
		}

		i := map[string]string{}
		r := new(models.Repository)

		if k == 0 {
			i["Tag"] = tag.(string)
		}
		i["id"] = image["id"].(string)

		//Put V1 JSON
		if err := r.PutJSONFromManifests(i, namespace, repository); err != nil {
			return err
		}

		if k == 0 {
			//Put V1 Tag
			if err := r.PutTagFromManifests(image["id"].(string), namespace, repository, tag.(string), string(data)); err != nil {
				return err
			}
		}

		img := new(models.Image)

		blobSum := manifest["fsLayers"].([]interface{})[k].(map[string]interface{})["blobSum"].(string)
		tarsum := strings.Split(blobSum, ":")[1]

		fmt.Println("[Registry API V2] Image %s sha256: %s", image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string))

		//Put Image Json
		if err := img.PutJSON(image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string), setting.APIVERSION_V2); err != nil {
			return err
		}

		//Put Image Layer
		basePath := setting.ImagePath
		layerfile := fmt.Sprintf("%v/uuid/%v/layer", basePath, tarsum)

		if err := img.PutLayer(image["id"].(string), layerfile, true, int64(image["Size"].(float64))); err != nil {
			return err
		}

		//Put Checksum
		if err := img.PutChecksum(image["id"].(string), tarsum, true, ""); err != nil {
			return err
		}

		//Put Ancestry
		if err := img.PutAncestry(image["id"].(string)); err != nil {
			return err
		}
	}

	return nil
}

func PutManifestsV2Handler(ctx *macaron.Context) (int, []byte) {

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	agent := ctx.Req.Header.Get("User-Agent")

	repo := new(models.Repository)
	if err := repo.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		result, _ := json.Marshal(map[string]string{"Error": err.Error()})
		return http.StatusBadRequest, result
	}

	manifest, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := manifestsConvertV1(manifest); err != nil {
		fmt.Errorf("[REGISTRY API V2] Decode Manifest Error: ", err.Error())
	}

	digest, err := DigestManifest(manifest)
	if err != nil {
		result, _ := json.Marshal(map[string]string{"Error": "Get manifest digest failure"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("http://%v/v2/%v/%v/manifests/%v",
		"containerops.me",
		namespace,
		repository,
		digest)
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusAccepted, []byte("")
}

func GetTagsListV2Handler(ctx *macaron.Context) (int, []byte) {
	result, _ := json.Marshal(map[string]string{"message": ""})

	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context) (int, []byte) {

	t := new(models.Tag)
	if err := t.Get(ctx.Params(":namespace"), ctx.Params(":repository"), ctx.Params(":tag")); err != nil {

		result, _ := json.Marshal(map[string]string{"Error": "Manifest not found"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")

	digest, err := DigestManifest([]byte(t.Manifest))
	if err != nil {
		result, _ := json.Marshal(map[string]string{"Error": "Get manifest digest failure"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}
