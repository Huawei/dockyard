package models

import (
	"encoding/json"
	"fmt"
	crew "github.com/containerops/crew/models"
	"github.com/containerops/wrench/db"
	"github.com/satori/go.uuid"
	"time"
)

func GetJSON(imageId string) (string, error) {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("Image not found")
	} else if !i.Checksumed || !i.Uploaded {
		return "", fmt.Errorf("Image JSON not found")
	} else {
		return i.JSON, nil
	}
}

func GetChecksum(imageId string) (string, error) {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("Image not found")
	} else if !i.Checksumed || !i.Uploaded {
		return "", fmt.Errorf("Image JSON not found")
	} else {
		return i.Checksum, nil
	}
}

func PutJSON(imageId, json string, version int64) error {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		i.ImageId = imageId
		i.UUID = string(db.GeneralDBKey(uuid.NewV4().String()))
		i.JSON = json
		i.Created = time.Now().UnixNano() / int64(time.Millisecond)
		i.Version = version

		fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "json:", json)

		if err = i.Save(); err != nil {
			return err
		}
	} else {
		i.ImageId, i.JSON = imageId, json
		i.Uploaded, i.Checksumed, i.Encrypted, i.Size, i.Updated, i.Version =
			false, false, false, 0, time.Now().UnixNano()/int64(time.Millisecond), version

		fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "json:", json)

		if err := i.Save(); err != nil {
			return err
		}
	}

	return nil
}

func PutChecksum(imageId string, checksum string, checksumed bool, payload string) error {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Image not found")
	} else {
		if err := PutAncestry(imageId); err != nil {

			return err
		}

		i.Checksum, i.Checksumed, i.Payload, i.Updated = checksum, checksumed, payload, time.Now().UnixNano()/int64(time.Millisecond)

		if err = i.Save(); err != nil {
			return err
		}

		if _, err := db.Client.HSet(db.GLOBAL_TARSUM_INDEX, checksum, i.UUID).Result(); err != nil {
			return err
		}

		fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "checksum:", i.Checksum)

	}

	return nil
}

func PutAncestry(imageId string) error {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Image not found")
	}

	var imageJSONMap map[string]interface{}
	var imageAncestry []string

	if err := json.Unmarshal([]byte(i.JSON), &imageJSONMap); err != nil {
		return err
	}

	if value, has := imageJSONMap["parent"]; has == true {
		parentImage := new(crew.Image)
		parentHas, _, err := parentImage.Has(value.(string))
		if err != nil {
			return err
		}

		if !parentHas {
			return fmt.Errorf("Parent image not found")
		}

		var parentAncestry []string
		json.Unmarshal([]byte(parentImage.Ancestry), &parentAncestry)
		imageAncestry = append(imageAncestry, imageId)
		imageAncestry = append(imageAncestry, parentAncestry...)
	} else {
		imageAncestry = append(imageAncestry, imageId)
	}

	ancestryJSON, _ := json.Marshal(imageAncestry)
	i.Ancestry = string(ancestryJSON)

	if err := i.Save(); err != nil {
		return err
	}

	fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "ancestry:", i.Ancestry)

	return nil
}

func PutLayer(imageId string, path string, uploaded bool, size int64) error {

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Image not found")
	} else {
		i.Path, i.Uploaded, i.Size, i.Updated = path, uploaded, size, time.Now().UnixNano()/int64(time.Millisecond)

		if err := i.Save(); err != nil {
			return err
		}

		fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "path:", path)
		fmt.Println("[REGISTRY API V1&V2]", i.ImageId, "size:", size)
	}

	return nil
}
