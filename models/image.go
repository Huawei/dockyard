package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/containerops/wrench/db"
)

type Image struct {
	Id         int64     `json:"id" orm:"auto"`
	ImageId    string    `json:"imageid" orm:"unique;varchar(255)"`
	JSON       string    `json:"json" orm:"null;type(text)"`
	Ancestry   string    `json:"ancestry" orm:"null;type(text)"`
	Checksum   string    `json:"checksum" orm:"null;varchar(255)"`
	Payload    string    `json:"payload" orm:"null;varchar(255)"`
	Checksumed bool      `json:"checksumed" orm:"null;default(0)"`
	Uploaded   bool      `json:"uploaded" orm:"null;default(0)"`
	Path       string    `json:"path" orm:"null;varchar(255)"`
	Size       int64     `json:"size" orm:"default(0)"`
	Version    int64     `json:"version" orm:"default(0)"`
	ManiPath   string    `json:"manipath" orm:"null;varchar(255)"`
	SignPath   string    `json:"signpath" orm:"null;varchar(255)"`
	AciPath    string    `json:"acipath" orm:"null;varchar(255)"`
	Memo       string    `json:"memo" orm:"null;varchar(255)"`
	Created    time.Time `json:"created" orm:"auto_now_add;type(datetime)"`
	Updated    time.Time `json:"updated" orm:"auto_now;type(datetime)"`
}

func (i *Image) Get(imageid string) (bool, error) {
	i.ImageId = imageid
	return db.Drv.Get(i, imageid)
}

func (i *Image) Save(imageid string) error {
	img := Image{ImageId: imageid}
	exists, err := img.Get(imageid)
	if err != nil {
		return err
	}

	i.ImageId = imageid
	if !exists {
		err = db.Drv.Insert(i)
	} else {
		err = db.Drv.Update(i)
	}

	return err
}

func (i *Image) GetJSON(imageId string) (string, error) {
	if exists, err := i.Get(imageId); err != nil {
		return "", err
	} else if exists == false {
		return "", fmt.Errorf("Not found image")
	} else if !i.Checksumed || !i.Uploaded {
		return "", fmt.Errorf("Not found image JSON")
	} else {
		return i.JSON, nil
	}
}

func (i *Image) GetPayload(imageId string) (string, error) {
	if exists, err := i.Get(imageId); err != nil {
		return "", err
	} else if exists == false {
		return "", fmt.Errorf("Not found image")
	} else if !i.Checksumed || !i.Uploaded {
		return "", fmt.Errorf("Not found image payload")
	} else {
		return i.Payload, nil
	}
}

func (i *Image) PutJSON(imageId, json string, version int64) error {
	if exists, err := i.Get(imageId); err != nil {
		return err
	} else if exists == false {
		i.ImageId = imageId
		i.JSON = json
		i.Version = version
	} else {
		i.ImageId = imageId
		i.JSON = json
		i.Uploaded = false
		i.Checksumed = false
		i.Size = 0
		i.Version = version
	}

	if err := i.Save(imageId); err != nil {
		return err
	}

	return nil
}

func (i *Image) PutLayer(imageId string, path string, uploaded bool, size int64) error {
	if exists, err := i.Get(imageId); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found image")
	} else {
		i.Path = path
		i.Uploaded = uploaded
		i.Size = size

		if err := i.Save(imageId); err != nil {
			return err
		}
	}

	return nil
}

func (i *Image) PutChecksum(imageId string, checksum string, checksumed bool, payload string) error {
	if exists, err := i.Get(imageId); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found image")
	} else {
		if err := i.PutAncestry(imageId); err != nil {

			return err
		}

		i.Checksum = checksum
		i.Checksumed = checksumed
		i.Payload = payload

		if err = i.Save(imageId); err != nil {
			return err
		}
	}

	return nil
}

func (i *Image) PutAncestry(imageId string) error {
	if exists, err := i.Get(imageId); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found image")
	}

	var imageJSONMap map[string]interface{}
	var imageAncestry []string
	if err := json.Unmarshal([]byte(i.JSON), &imageJSONMap); err != nil {
		return err
	}

	if value, has := imageJSONMap["parent"]; has == true {
		parentImage := new(Image)
		parentHas, err := parentImage.Get(value.(string))
		if err != nil {
			return err
		}

		if !parentHas {
			return fmt.Errorf("Not found parent image")
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

	if err := i.Save(imageId); err != nil {
		return err
	}

	return nil
}
