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
	UpMan     bool   `json:"upload_manifest_mark"`
	UpSig     bool   `json:"upload_signature_mark"`
	UpAci     bool   `json:"upload_aci_mark"`
}

type UploadDetails struct {
	ACIPushVersion string `json:"aci_push_version"`
	Multipart      bool   `json:"multipart"`
	ManifestURL    string `json:"upload_manifest_url"`
	SignatureURL   string `json:"upload_signature_url"`
	ACIURL         string `json:"upload_aci_url"`
	CompletedURL   string `json:"completed_url"`
}

type CompleteMsg struct {
	Success      bool   `json:"success"`
	Reason       string `json:"reason,omitempty"`
	ServerReason string `json:"server_reason,omitempty"`
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

	//TODO:repository is empry when acpush first time connect with server. need acpush`s coorpation.

	if err := db.Get(r, key); err != nil {
		return err
	}

	return nil
}

func (r *AciRepository) AciIsExisted(namespace string, imgname string) (bool, error) {
	if err := r.GetRepository(namespace); err != nil {
		return false, err
	}

	for _, aci := range r.Acis {
        if aci.ImageName == imgname {
			return true,nil
        }
	}

	return false, fmt.Errorf("Invalid repository key")
}

func (r *AciRepository) GetAciByName(namespace string, imgname string) (*AciDetail, error) {
	if err := r.GetRepository(namespace); err != nil {
		return nil, err
	}

    if len(r.Acis) > 0 {
		for _, aciDst := range r.Acis {
	        if aciDst.ImageName == imgname {
				return &aciDst,nil
	        }
		}
	} else {
        return nil, fmt.Errorf("Acis of repository is empty")
	}

	return nil, fmt.Errorf("Invalid repository key")
}

func (r *AciRepository) PutManifest(namespace string, imgname string, manifest string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	if err := db.Get(r, key); err != nil {
		return err
	}

    if b, _ := r.AciIsExisted(namespace, imgname); b == true { 
		for i, aci := range r.Acis {
		    if aci.ImageName == imgname {
		        r.Acis[i].Manifest = manifest
		        r.Acis[i].UpMan    = true
		    }
		}
	} else {
    	r.Acis = append(r.Acis, AciDetail{
			AciID:     "",
			ImageName: imgname,
			Manifest:  manifest,
			SignPath:  "",
			AciPath:   "",
			UpMan:     true,
			UpSig:     false,
			UpAci:     false,			
		})
	}

	if err := db.Save(r, key); err != nil {
		return err
	}
	return nil
}

func (r *AciRepository) PutSignpath(namespace string, imgname string, signpath string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	if err := db.Get(r, key); err != nil {
		return err
	}

    if len(r.Acis) > 0 {
		for i, aci := range r.Acis {
		    if aci.ImageName == imgname {
		        r.Acis[i].SignPath = signpath
		        r.Acis[i].UpSig    = true
		    }
		}
	} else {
       return fmt.Errorf("Acis of repository is empty")
	}

	if err := db.Save(r, key); err != nil {
		return err
	}
	return nil
}

func (r *AciRepository) PutAcipath(namespace string, imgname string, acipath string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	if err := db.Get(r, key); err != nil {
		return err
	}

    if len(r.Acis) > 0 {
		for i, aci := range r.Acis {
		    if aci.ImageName == imgname {
		        r.Acis[i].AciPath  = acipath
		        r.Acis[i].UpAci    = true
		    }
		}
	} else {
       return fmt.Errorf("Acis of repository is empty")
	}

	if err := db.Save(r, key); err != nil {
		return err
	}
	return nil
}
