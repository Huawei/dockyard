package modules

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/utils"
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

		if err := r.PutJSONFromManifests(i, namespace, repository); err != nil {
			return err
		}

		if k == 0 {
			if err := r.PutTagFromManifests(image["id"].(string), namespace, repository, tag.(string), string(data)); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyImgLayer(srcPath, srcFile, dstPath, dstFile string, resp io.Reader) (int, error) {
	if !utils.IsDirExist(dstPath) {
		os.MkdirAll(dstPath, os.ModePerm)
	}

	if utils.IsFileExist(dstFile) {
		os.Remove(dstFile)
	}

	var data []byte
	if _, err := os.Stat(srcFile); err == nil {
		data, _ = ioutil.ReadFile(srcFile)
		if err := ioutil.WriteFile(dstFile, data, 0777); err != nil {
			return 0, err
		}
		os.RemoveAll(srcPath)
	} else {
		data, _ = ioutil.ReadAll(resp)
		if err := ioutil.WriteFile(dstFile, data, 0777); err != nil {
			return 0, err
		}
	}

	return len(data), nil
}
