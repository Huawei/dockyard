package models

import (
	"github.com/containerops/wrench/setting"
)

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
	ServerReason string `json:"serverreason,omitempty"`
}

type TemplateDesc struct {
	NameSpace  string
	AciName    string
	Domains    string
	ListenMode string
}

var TemplatePath string = "views/aci/index.html"
var AcipathExist bool = true

func (r *Repository) Update(namespace, aciname, aciid, maniPath, signPath, aciPath string) error {
	has, _, err := r.Has(namespace, aciname)
	if err != nil {
		return err
	}

	r.Aci.AciID, r.Aci.AciName, r.Aci.ManiPath, r.Aci.SignPath, r.Aci.AciPath =
		aciid, aciname, maniPath, signPath, aciPath

	if has == true {
		if err := r.Save(); err != nil {
			return err
		}
	} else {
		if err := r.Put(namespace, aciname, "", "", setting.APIVERSION_ACI); err != nil {
			return err
		}
	}

	return nil
}

/*
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

func (r *AciRepository) CreateRepository(namespace string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	r = &AciRepository{
		NameSpace: namespace,
	}

	if err := db.Save(r, key); err != nil {
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
			return true, nil
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
				return &aciDst, nil
			}
		}
	} else {
		return nil, fmt.Errorf("Acis of repository is empty")
	}

	return nil, fmt.Errorf("can`t get currect aci in %v repository by name:%v ", namespace, imgname)
}

func (r *AciRepository) PutAciByName(namespace string, imgname string, signpath string, acipath string, keyspath string) error {
	key := db.Key("repository", namespace, "")
	if len(key) <= 0 {
		return fmt.Errorf("Invalid repository key")
	}

	//create a new repository and load it when user pushs acis at first time
	if err := db.Get(r, key); err != nil {
		if err := r.CreateRepository(namespace); err != nil {
			return fmt.Errorf("Create repository fail")
		}
	}

	if b, _ := r.AciIsExisted(namespace, imgname); b == true {
		for i, aci := range r.Acis {
			if aci.ImageName == imgname {
				r.Acis[i].SignPath = signpath
				r.Acis[i].AciPath = acipath
			}
		}
	} else {
		r.Acis = append(r.Acis, AciDetail{
			ImageName: imgname,
			SignPath:  signpath,
			AciPath:   acipath,
		})
	}
	r.PubKeysPath = keyspath

	if err := db.Save(r, key); err != nil {
		return err
	}
	return nil
}
*/
