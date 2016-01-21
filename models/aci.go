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
	Namespace  string
	Repository string
	Domains    string
	ListenMode string
}

type Aci struct{}

var TemplatePath string = "views/aci/index.html"

//TODO: to consider the situation that new ACI repository instead Docker repository
func (a *Aci) Update(namespace, repository, tag, imageId, manipath, signpath, acipath string) error {
	r := new(Repository)
	r.Namespace, r.Repository, r.Version = namespace, repository, setting.APIVERSION_ACI
	if err := r.Save(); err != nil {
		return err
	}

	i := new(Image)
	i.ImageId, i.ManiPath, i.SignPath, i.AciPath = imageId, manipath, signpath, acipath
	if err := i.Save(); err != nil {
		return err
	}

	t := new(Tag)
	t.Name, t.Namespace, t.Repository, t.ImageId = tag, namespace, repository, imageId
	if err := t.Save(); err != nil {
		return err
	}

	return nil
}
