package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/redis.v3"

	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

type Repository struct {
	Repository    string   `json:"repository"`    //
	Namespace     string   `json:"namespace"`     //
	NamespaceType bool     `json:"namespacetype"` //
	Organization  string   `json:"organization"`  //
	Tags          []string `json:"tags"`          //
	Starts        []string `json:"starts"`        //
	Comments      []string `json:"comments"`      //
	Short         string   `json:"short"`         //
	Description   string   `json:"description"`   //
	JSON          string   `json:"json"`          //
	Dockerfile    string   `json:"dockerfile"`    //
	Agent         string   `json:"agent"`         //
	Links         string   `json:"links"`         //
	Size          int64    `json:"size"`          //
	Download      int64    `json:"download"`      //
	Uploaded      bool     `json:"uploaded"`      //
	Checksum      string   `json:"checksum"`      //
	Checksumed    bool     `json:"checksumed"`    //
	Icon          string   `json:"icon"`          //
	Sign          string   `json:"sign"`          //
	Privated      bool     `json:"privated"`      //
	Clear         string   `json:"clear"`         //
	Cleared       bool     `json:"cleared"`       //
	Encrypted     bool     `json:"encrypted"`     //
	Created       int64    `json:"created"`       //
	Updated       int64    `json:"updated"`       //
	Version       int64    `json:"version"`       //
	Memo          []string `json:"memo"`          //
	Aci           AciDesc  `json:"aci"`           //
}

type Tag struct {
	Name       string   `json:"name"`       //
	ImageId    string   `json:"imageid"`    //
	Namespace  string   `json:"namespace"`  //
	Repository string   `json:"repository"` //
	Sign       string   `json:"sign"`       //
	Manifest   string   `json:"manifest"`   //
	Memo       []string `json:"memo"`       //
}

type AciDesc struct {
	AciID    string `json:"aciid"`
	AciName  string `json:"aciname"`
	ManiPath string `json:"manipath"`
	SignPath string `json:"signpath"`
	AciPath  string `json:"acipath"`
}

func (r *Repository) Has(namespace, repository string) (bool, string, error) {
	if key := db.Key("repository", namespace, repository); len(key) <= 0 {
		return false, "", fmt.Errorf("Invalid repository key")
	} else {
		if err := db.Get(r, key); err != nil {
			if err == redis.Nil {
				return false, "", nil
			} else {
				return false, "", err
			}
		}

		return true, key, nil
	}
}

func (r *Repository) Save() error {
	key := db.Key("repository", r.Namespace, r.Repository)

	if err := db.Save(r, key); err != nil {
		return err
	}

	if _, err := db.Client.HSet(db.GLOBAL_REPOSITORY_INDEX, (fmt.Sprintf("%s/%s", r.Namespace, r.Repository)), key).Result(); err != nil {
		return err
	}

	return nil
}

func (t *Tag) Save() error {
	key := db.Key("tag", t.Namespace, t.Repository, t.Name)

	if err := db.Save(t, key); err != nil {
		return err
	}

	if _, err := db.Client.HSet(db.GLOBAL_TAG_INDEX, (fmt.Sprintf("%s/%s/%s:%s", t.Namespace, t.Repository, t.Name, t.ImageId)), key).Result(); err != nil {
		return err
	}

	return nil
}

func (t *Tag) Get(namespace, repository, tag string) error {
	key := db.Key("tag", namespace, repository, tag)

	if err := db.Get(t, key); err != nil {
		return err
	}

	return nil
}

func (t *Tag) GetByKey(key string) error {
	if err := db.Get(t, key); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Put(namespace, repository, json, agent string, version int64) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		r.Created = time.Now().UnixNano() / int64(time.Millisecond)
	}

	r.Namespace, r.Repository, r.JSON, r.Agent, r.Version =
		namespace, repository, json, agent, version

	r.Updated = time.Now().UnixNano() / int64(time.Millisecond)
	r.Checksumed, r.Uploaded, r.Cleared, r.Encrypted = false, false, false, false
	r.Size, r.Download = 0, 0

	if err := r.Save(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutImages(namespace, repository string) error {
	if _, _, err := r.Has(namespace, repository); err != nil {
		return err
	}

	r.Checksumed, r.Uploaded, r.Updated = true, true, time.Now().Unix()

	if err := r.Save(); err != nil {
		return fmt.Errorf("[REGISTRY API V1] Update Uploaded flag error")
	}

	return nil
}

func (r *Repository) PutTag(imageId, namespace, repository, tag string) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Repository not found")
	}

	i := new(Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Tag's image not found")
	}

	t := new(Tag)
	t.Name, t.ImageId, t.Namespace, t.Repository = tag, imageId, namespace, repository

	if err := t.Save(); err != nil {
		return err
	}

	has := false
	for _, value := range r.Tags {
		if value == db.Key("tag", t.Namespace, t.Repository, t.Name) {
			has = true
		}
	}

	if !has {
		r.Tags = append(r.Tags, db.Key("tag", t.Namespace, t.Repository, t.Name))
	}

	if err := r.Save(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutJSONFromManifests(image map[string]string, namespace, repository string) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		r.Created = time.Now().UnixNano() / int64(time.Millisecond)
		r.JSON = ""
	}

	r.Namespace, r.Repository, r.Version = namespace, repository, setting.APIVERSION_V2

	r.Updated = time.Now().UnixNano() / int64(time.Millisecond)
	r.Checksumed, r.Uploaded, r.Cleared, r.Encrypted = true, true, true, false
	r.Size, r.Download = 0, 0

	if len(r.JSON) == 0 {
		if data, err := json.Marshal([]map[string]string{image}); err != nil {
			return err
		} else {
			r.JSON = string(data)
		}

	} else {
		var ids []map[string]string

		if err := json.Unmarshal([]byte(r.JSON), &ids); err != nil {
			return err
		}

		has := false
		for _, v := range ids {
			if v["id"] == image["id"] {
				has = true
			}
		}

		if has == false {
			ids = append(ids, image)
		}

		if data, err := json.Marshal(ids); err != nil {
			return err
		} else {
			r.JSON = string(data)
		}
	}

	if err := r.Save(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutTagFromManifests(image, namespace, repository, tag, manifests string) error {
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Repository not found")
	}

	t := new(Tag)
	t.Name, t.ImageId, t.Namespace, t.Repository, t.Manifest = tag, image, namespace, repository, manifests

	if err := t.Save(); err != nil {
		return err
	}

	has := false
	for _, v := range r.Tags {
		if v == db.Key("tag", t.Namespace, t.Repository, t.Name) {
			has = true
		}
	}

	if has == false {
		r.Tags = append(r.Tags, db.Key("tag", t.Namespace, t.Repository, t.Name))
	}

	if err := r.Save(); err != nil {
		return err
	}

	return nil
}
