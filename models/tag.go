package models

import (
	"time"

	"github.com/containerops/dockyard/utils/db"
)

type Tag struct {
	Id         int64     `json:"id" orm:"auto"`
	Namespace  string    `json:"namespace" orm:"varchar(255)"`
	Repository string    `json:"repository" orm:"varchar(255)"`
	Tag        string    `json:"tag" orm:"varchar(255)"`
	ImageId    string    `json:"imageid" orm:"varchar(255)"`
	Manifest   string    `json:"manifest" orm:"null;type(text)"`
	Schema     int64     `json:"schema" orm:"default(0)"`
	Memo       string    `json:"memo" orm:"null;type(text)"`
	Created    time.Time `json:"created" orm:"auto_now_add;type(datetime)"`
	Updated    time.Time `json:"updated" orm:"auto_now;type(datetime)"`
}

func (t *Tag) TableUnique() [][]string {
	return [][]string{
		[]string{"Namespace", "Repository", "Tag"},
	}
}

func (t *Tag) Get(namespace, repository, tag string) (bool, error) {
	t.Namespace, t.Repository, t.Tag = namespace, repository, tag
	return db.Drv.Get(t, namespace, repository, tag)
}

func (t *Tag) Save(namespace, repository, tag string) error {
	tg := Tag{Namespace: namespace, Repository: repository, Tag: tag}
	exists, err := tg.Get(namespace, repository, tag)
	if err != nil {
		return err
	}

	t.Namespace, t.Repository, t.Tag = namespace, repository, tag
	if !exists {
		err = db.Drv.Insert(t)
	} else {
		err = db.Drv.Update(t)
	}

	return err

}
