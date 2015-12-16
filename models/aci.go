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

type UploadDetails struct {
	ACIPushVersion string `json:"aci_push_version"`
	Multipart      bool   `json:"multipart"`
	ManifestURL    string `json:"upload_manifest_url"`
	SignatureURL   string `json:"upload_signature_url"`
	ACIURL         string `json:"upload_aci_url"`
	CompletedURL   string `json:"completed_url"`
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

func (r *AciRepository) GetAciByName(namespace string, imgname string) (*AciDetail, error) {
	if err := r.GetRepository(namespace); err != nil {
		return nil, err
	}

	for _, aci := range r.Acis {
        if aci.ImageName == imgname {
			return &aci,nil
        }
	}

	return nil,fmt.Errorf("Invalid repository key")
}

func (r *AciRepository) PutAciByName(namespace string, imgname string, aciSrc *AciDetail) error {
	if err := r.GetRepository(namespace); err != nil {
		return err
	}

	for _, aci := range r.Acis {
        if aci.ImageName == imgname {
			aci = *aciSrc
			return nil
        }
	}

	return fmt.Errorf("Invalid repository key")
}
