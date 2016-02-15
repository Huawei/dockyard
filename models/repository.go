package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/setting"
)

type Repository struct {
	Id         int64     `json:"id" orm:"auto"`
	Namespace  string    `json:"namespace" orm:"varchar(255)"`
	Repository string    `json:"repository" orm:"varchar(255)"`
	JSON       string    `json:"json" orm:"null;type(text)"`
	Agent      string    `json:"agent" orm:"null;type(text)"`
	Size       int64     `json:"size" orm:"default(0)"`
	Version    int64     `json:"version" orm:"default(0)"`
	Tagslist   string    `json:"tagslist" orm:"null;type(text)"`
	Download   int64     `json:"download" orm:"null;default(0)"`
	Memo       string    `json:"memo" orm:"null;varchar(255)"`
	Created    time.Time `json:"created" orm:"auto_now_add;type(datetime)"`
	Updated    time.Time `json:"updated" orm:"auto_now;type(datetime)"`
}

func (r *Repository) TableUnique() [][]string {
	return [][]string{
		[]string{"Namespace", "Repository"},
	}
}

func (r *Repository) Get(namespace, repository string) (bool, error) {
	r.Namespace, r.Repository = namespace, repository
	return db.Drv.Get(r, namespace, repository)
}

func (r *Repository) Save(namespace, repository string) error {
	rep := Repository{Namespace: namespace, Repository: repository}
	exists, err := rep.Get(namespace, repository)
	if err != nil {
		return err
	}

	r.Namespace, r.Repository = namespace, repository
	if !exists {
		err = db.Drv.Insert(r)
	} else {
		err = db.Drv.Update(r)
	}

	return err
}

func (r *Repository) GetTagslist() []string {
	if len(r.Tagslist) <= 0 {
		return []string{}
	}

	return strings.Split(r.Tagslist, ",")
}

func (r *Repository) SaveTagslist(tagslist []string) string {
	if len(tagslist) <= 0 {
		return ""
	}

	return strings.Join(tagslist, ",")
}

func (r *Repository) PutJSONFromManifests(image map[string]string, namespace, repository string) error {
	if exists, err := r.Get(namespace, repository); err != nil {
		return err
	} else if exists == false {
		r.JSON = ""
	}

	r.Namespace = namespace
	r.Repository = repository
	r.Version = setting.APIVERSION_V2
	r.Size = 0
	r.Download = 0

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

		exists := false
		for _, v := range ids {
			if v["id"] == image["id"] {
				exists = true
			}
		}

		if exists == false {
			ids = append(ids, image)
		}

		if data, err := json.Marshal(ids); err != nil {
			return err
		} else {
			r.JSON = string(data)
		}
	}

	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutTagFromManifests(image, namespace, repository, tag, manifests string, schema int64) error {
	if exists, err := r.Get(namespace, repository); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found repository")
	}

	t := new(Tag)
	t.Tag = tag
	t.ImageId = image
	t.Namespace = namespace
	t.Repository = repository
	t.Manifest = manifests
	t.Schema = schema

	if err := t.Save(t.Namespace, t.Repository, t.Tag); err != nil {
		return err
	}

	exists := false
	tagslist := r.GetTagslist()
	for _, v := range tagslist {
		if v == t.Tag {
			exists = true
			break
		}
	}

	if !exists {
		tagslist = append(tagslist, t.Tag)
	}

	r.Tagslist = r.SaveTagslist(tagslist)

	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	return nil
}

func (r *Repository) Put(namespace, repository, json, agent string, version int64) error {
	if _, err := r.Get(namespace, repository); err != nil {
		return err
	}

	r.Namespace = namespace
	r.Repository = repository
	r.JSON = json
	r.Agent = agent
	r.Version = version
	r.Size = 0
	r.Download = 0

	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutTag(imageId, namespace, repository, tag string) error {
	if exists, err := r.Get(namespace, repository); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found  repository")
	}

	i := new(Image)
	if exists, err := i.Get(imageId); err != nil {
		return err
	} else if exists == false {
		return fmt.Errorf("Not found Tag's image")
	}

	t := new(Tag)
	t.Tag, t.ImageId, t.Namespace, t.Repository = tag, imageId, namespace, repository

	if err := t.Save(namespace, repository, tag); err != nil {
		return err
	}

	exists := false
	tagslist := r.GetTagslist()
	for _, v := range tagslist {
		if v == t.Tag {
			exists = true
			break
		}
	}

	if !exists {
		tagslist = append(tagslist, t.Tag)
	}

	r.Tagslist = r.SaveTagslist(tagslist)

	if err := r.Save(namespace, repository); err != nil {
		return err
	}

	return nil
}

func (r *Repository) PutImages(namespace, repository string) error {
	if exists, err := r.Get(namespace, repository); err != nil {
		return err
	} else if !exists {
		if err := r.Save(namespace, repository); err != nil {
			return err
		}
	}

	//r.Checksumed, r.Uploaded, r.Updated = true, true, time.Now().Unix()
	return nil
}
