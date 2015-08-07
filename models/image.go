package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/redis.v3"

	"github.com/containerops/wrench/db"
)

type Image struct {
	ImageId    string   `json:"imageid"`    //
	JSON       string   `json:"json"`       //
	Ancestry   string   `json:"ancestry"`   //
	Checksum   string   `json:"checksum"`   // tarsum+sha256
	Payload    string   `json:"payload"`    // sha256
	URL        string   `json:"url"`        //
	Backend    string   `json:"backend"`    //
	Path       string   `json:"path"`       //
	Sign       string   `json:"sign"`       //
	Size       int64    `json:"size"`       //
	Uploaded   bool     `json:"uploaded"`   //
	Checksumed bool     `json:"checksumed"` //
	Encrypted  bool     `json:"encrypted"`  //
	Created    int64    `json:"created"`    //
	Updated    int64    `json:"updated"`    //
	Memo       []string `json:"memo"`       //
	Version    int64    `json:"version"`    //
}

func (i *Image) Has(image string) (bool, string, error) {
	if key := db.Key("image", image); len(key) <= 0 {
		return false, "", fmt.Errorf("Invalid image key")
	} else {
		if err := db.Get(i, key); err != nil {
			if err == redis.Nil {
				return false, "", nil
			} else {
				return false, "", err
			}
		}

		return true, key, nil
	}
}

func (i *Image) Save() error {
	key := db.Key("image", i.ImageId)

	if err := db.Save(i, key); err != nil {
		return err
	}

	if _, err := db.Client.HSet(db.GLOBAL_IMAGE_INDEX, i.ImageId, key).Result(); err != nil {
		return err
	}

	return nil
}

func (i *Image) GetJSON(imageId string) (string, error) {
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

func (i *Image) GetChecksumPayload(imageId string) (string, error) {

	if has, _, err := i.Has(imageId); err != nil {
		return "", err
	} else if has == false {
		return "", fmt.Errorf("Image not found")
	} else if !i.Checksumed || !i.Uploaded {
		return "", fmt.Errorf("Image JSON not found")
	} else {
		return i.Payload, nil
	}
}

func (i *Image) PutJSON(imageId, json string, version int64) error {

	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		i.ImageId = imageId
		i.JSON = json
		i.Created = time.Now().UnixNano() / int64(time.Millisecond)
		i.Version = version

		if err = i.Save(); err != nil {
			return err
		}
	} else {
		i.ImageId, i.JSON = imageId, json
		i.Uploaded, i.Checksumed, i.Encrypted, i.Size, i.Updated, i.Version =
			false, false, false, 0, time.Now().UnixNano()/int64(time.Millisecond), version

		if err := i.Save(); err != nil {
			return err
		}
	}

	return nil
}

func (i *Image) PutChecksum(imageId string, checksum string, checksumed bool, payload string) error {

	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Image not found")
	} else {
		if err := i.PutAncestry(imageId); err != nil {

			return err
		}

		i.Checksum, i.Checksumed, i.Payload, i.Updated = checksum, checksumed, payload, time.Now().UnixNano()/int64(time.Millisecond)

		if err = i.Save(); err != nil {
			return err
		}
	}

	return nil
}

func (i *Image) PutAncestry(imageId string) error {

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
		parentImage := new(Image)
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

	return nil
}

func (i *Image) PutLayer(imageId string, path string, uploaded bool, size int64) error {

	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Image not found")
	} else {
		i.Path, i.Uploaded, i.Size, i.Updated = path, uploaded, size, time.Now().UnixNano()/int64(time.Millisecond)

		if err := i.Save(); err != nil {
			return err
		}
	}

	return nil
}

func (i *Image) HasTarsum(tarsum string) (bool, []string, error) {
	if value, err := db.Client.HGet(db.GLOBAL_TARSUM_INDEX, tarsum).Result(); err != nil {
		return false, []string{}, err
	} else {
		var imggrp = []string{}
		if err := json.Unmarshal([]byte(value), &imggrp); err != nil {
			return false, []string{}, err
		}

		return true, imggrp, nil
	}
}

func (i *Image) PutTarsum(imageId, tarsum string) error {
	var imggrp = []string{}

	if value, err := db.Client.HGet(db.GLOBAL_TARSUM_INDEX, tarsum).Result(); err != nil {
		imggrp = append(imggrp, imageId)
	} else {
		if err := json.Unmarshal([]byte(value), &imggrp); err != nil {
			return err
		}

		for _, v := range imggrp {
			if v == imageId {
				return nil
			}
		}

		imggrp = append(imggrp, imageId)
	}

	result, _ := json.Marshal(imggrp)

	if _, err := db.Client.HSet(db.GLOBAL_TARSUM_INDEX, tarsum, string(result)).Result(); err != nil {
		return err
	}

	return nil
}
