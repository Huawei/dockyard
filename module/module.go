package modules

import (
	"encoding/json"
	"strings"

	"github.com/containerops/dockyard/models"
)

func ParseManifest(data []byte) error {

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
		/*
			img := new(models.Image)

			blobSum := manifest["fsLayers"].([]interface{})[k].(map[string]interface{})["blobSum"].(string)
			tarsum := strings.Split(blobSum, ":")[1]

			fmt.Println("[Registry API V2] Image %s sha256: %s", image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string))

			//Put Image Json
			if err := img.PutJSON(image["id"].(string), v.(map[string]interface{})["v1Compatibility"].(string), setting.APIVERSION_V2); err != nil {
				return err
			}

			//Put Image Layer
			basePath := setting.BasePath
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
		*/
	}

	return nil
}
