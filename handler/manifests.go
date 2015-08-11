package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Unknwon/macaron"
	"github.com/astaxie/beego/logs"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
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

		//log.Debug("[Registry API V2] Image %s sha256: %s", image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string))

		//Put Image Json
		if err := img.PutJSON(image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string), setting.APIVERSION_V2); err != nil {
			return err
		}

		//Put Image Layer,Compatible with V1,save the layerfile by imageId as the same with V1,and remove the temporary one
		basePath := setting.ImagePath
		layerfileTmp := fmt.Sprintf("%v/temp/%v/layer", basePath, tarsum)
		layerfile := fmt.Sprintf("%v/images/%v/layer", basePath, image["id"].(string))
		if _, err := os.Stat(layerfileTmp); err != nil {
			if !utils.IsFileExist(layerfile) {
				return err
			}
		} else {
			imagePath := fmt.Sprintf("%v/images/%v", setting.ImagePath, image["id"].(string))
			if !utils.IsDirExist(imagePath) {
				os.MkdirAll(imagePath, os.ModePerm)
			}

			data, _ := ioutil.ReadFile(layerfileTmp)
			if err := ioutil.WriteFile(layerfile, data, 0777); err != nil {
				return err
			}
		}

		if err := img.PutLayer(image["id"].(string), layerfile, true, int64(image["Size"].(float64))); err != nil {
			return err
		}

		//Put Checksum
		if err := img.PutChecksum(image["id"].(string), tarsum, true, ""); err != nil {
			return err
		}

		//Put V2 tarsum
		if err := img.PutTarsum(image["id"].(string), tarsum); err != nil {
			return err
		}

		//Put Ancestry
		if err := img.PutAncestry(image["id"].(string)); err != nil {
			return err
		}
	}

	os.RemoveAll(fmt.Sprintf("%v/temp", setting.ImagePath))
	return nil
}

func PutManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	agent := ctx.Req.Header.Get("User-Agent")

	repo := new(models.Repository)
	if err := repo.Put(namespace, repository, "", agent, setting.APIVERSION_V2); err != nil {
		log.Error("[REGISTRY API V2] Save repository failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": err.Error()})
		return http.StatusBadRequest, result
	}

	manifest, _ := ioutil.ReadAll(ctx.Req.Request.Body)
	if err := manifestsConvertV1(manifest); err != nil {
		log.Error("[REGISTRY API V2] Convert V2 manifests to V1 format failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Convert V2 manifests to V1 format failed"})
		return http.StatusBadRequest, result
	}

	digest, err := DigestManifest(manifest)
	if err != nil {
		log.Error("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get manifest digest failed"})
		return http.StatusBadRequest, result
	}

	random := fmt.Sprintf("https://%v/v2/%v/%v/manifests/%v",
		setting.Domains,
		namespace,
		repository,
		digest)
	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Location", random)

	return http.StatusAccepted, []byte("")
}

func GetTagsListV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	r := new(models.Repository)
	if has, _, err := r.Has(namespace, repository); err != nil || has == false {
		log.Error("[REGISTRY API V2] Repository not found: %v", repository)

		result, _ := json.Marshal(map[string]string{"message": "Repository not found"})
		return http.StatusNotFound, result
	}

	data := map[string]interface{}{}
	tags := []string{}
	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	for _, value := range r.Tags {
		t := new(models.Tag)
		if err := t.GetByKey(value); err != nil {
			log.Error("[REGISTRY API V2] Tag not found: %v", err.Error())

			result, _ := json.Marshal(map[string]string{"message": "Tag not found"})
			return http.StatusNotFound, result
		}

		tags = append(tags, t.Name)
	}

	data["tags"] = tags

	result, _ := json.Marshal(data)
	return http.StatusOK, result
}

func GetManifestsV2Handler(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	t := new(models.Tag)
	if err := t.Get(ctx.Params(":namespace"), ctx.Params(":repository"), ctx.Params(":tag")); err != nil {
		log.Error("[REGISTRY API V2] Manifest not found: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Manifest not found"})
		return http.StatusNotFound, result
	}

	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")

	digest, err := DigestManifest([]byte(t.Manifest))
	if err != nil {
		log.Error("[REGISTRY API V2] Get manifest digest failed: %v", err.Error())

		result, _ := json.Marshal(map[string]string{"message": "Get manifest digest failed"})
		return http.StatusBadRequest, result
	}

	ctx.Resp.Header().Set("Docker-Content-Digest", digest)
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(t.Manifest)))

	return http.StatusOK, []byte(t.Manifest)
}
