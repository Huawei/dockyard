package models

import (
	"fmt"
	"github.com/Unknwon/macaron"
	crew "github.com/containerops/crew/models"
	"github.com/containerops/wrench/db"
	"github.com/containerops/wrench/utils"
	"time"
)

func Put(namespace, repository, json, agent string, version int64) error {
	r := new(crew.Repository)
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		r.UUID = db.GeneralDBKey(fmt.Sprintf("%s:%s", namespace, repository))
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

func PutImages(namespace, repository string, ctx *macaron.Context) error {

	r := new(crew.Repository)
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("[REGISTRY API V1] Repository not found")
	}

	r.Checksumed, r.Uploaded, r.Updated = true, true, time.Now().Unix()

	if err := r.Save(); err != nil {
		return fmt.Errorf("[REGISTRY API V1] Update Uploaded flag error")
	}

	org := new(crew.Organization)
	isOrg, _, err := org.Has(namespace)
	if err != nil {
		return fmt.Errorf("[REGISTRY API V1] Search Organization Error")
	}

	user := new(crew.User)
	authUsername, _, _ := utils.DecodeBasicAuth(ctx.Req.Header.Get("Authorization"))
	isUser, _, err := user.Has(authUsername)
	if err != nil {
		return fmt.Errorf("[REGISTRY API V1] Search User Error")
	}

	if !isUser && !isOrg {
		return fmt.Errorf("[REGISTRY API V1] Search Namespace Error")
	}

	if isUser {
		user.Repositories = append(user.Repositories, r.UUID)
		user.Save()
	}
	if isOrg {
		org.Repositories = append(org.Repositories, r.UUID)
		org.Save()
	}

	return nil
}

func PutTag(imageId, namespace, repository, tag string) error {
	r := new(crew.Repository)
	if has, _, err := r.Has(namespace, repository); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Repository not found")
	}

	i := new(crew.Image)
	if has, _, err := i.Has(imageId); err != nil {
		return err
	} else if has == false {
		return fmt.Errorf("Tag's image not found")
	}

	t := new(crew.Tag)
	t.UUID = string(fmt.Sprintf("%s:%s:%s", namespace, repository, tag))
	t.Name, t.ImageId, t.Namespace, t.Repository = tag, imageId, namespace, repository

	if err := t.Save(); err != nil {
		return err
	}

	has := false
	for _, value := range r.Tags {
		if value == t.UUID {
			has = true
		}
	}
	if !has {
		r.Tags = append(r.Tags, t.UUID)
	}
	if err := r.Save(); err != nil {
		return err
	}

	return nil
}
