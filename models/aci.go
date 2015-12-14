package models

import (
	"fmt"

	"github.com/containerops/wrench/db"
)

type AciRepository struct {
	NameSpace   string      `json:"namesapce"`
	Acis        []AciDetail `json:"acis"`
	PubKeysPath string      `json:"pubKeyspath"`
}

type AciDetail struct {
	AciID     string `json:"aciid"`
	ImageName string `json:"imagename"`
	Manifest  string `json:"manifest"`
	SignPath  string `json:"signpath"`
	AciPath   string `json:"acipath"`
}

type TemplateDesc struct {
	NameSpace  string
	AciName    string
	ServerName string
	ListenMode string
}

var TemplatePath string = "views/aci/index.html"

func (r *AciRepository) GetRepository(namespace string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	if err := db.Get(r, key); err != nil {
		return err
	}

	return nil
}
