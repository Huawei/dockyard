package models

import (
	"fmt"
	"encoding/json"

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

type CompleteMsg struct {
	Success      bool   `json:"success"`
	Reason       string `json:"reason,omitempty"`
	ServerReason string `json:"server_reason,omitempty"`
}

//aci manifest struct
type ImageManifest struct {
	ACKind        string             `json:"acKind"`
	ACVersion     string             `json:"acVersion"`
	Name          string             `json:"name"`
	Labels        []Label            `json:"labels,omitempty"`
	App           *App               `json:"app,omitempty"`
	Annotations   []Annotation       `json:"annotations,omitempty"`
	Dependencies  []Dependency       `json:"dependencies,omitempty"`
	PathWhitelist []string           `json:"pathWhitelist,omitempty"`
}

type Version struct {
	Major      int64
	Minor      int64
	Patch      int64
	PreRelease string
	Metadata   string
}

type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type App struct {
	Exec              []string       `json:"exec"`
	EventHandlers     []EventHandler `json:"eventHandlers,omitempty"`
	User              string         `json:"user"`
	Group             string         `json:"group"`
	SupplementaryGIDs []int          `json:"supplementaryGIDs,omitempty"`
	WorkingDirectory  string         `json:"workingDirectory,omitempty"`
	Environment       []EnvironmentVariable  `json:"environment,omitempty"`
	MountPoints       []MountPoint   `json:"mountPoints,omitempty"`
	Ports             []Port         `json:"ports,omitempty"`
	Isolators         []Isolator     `json:"isolators,omitempty"`
}

type EventHandler struct {
	Name string     `json:"name"`
	Exec []string   `json:"exec"`
}

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MountPoint struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

type Port struct {
	Name            string `json:"name"`
	Protocol        string `json:"protocol"`
	Port            uint   `json:"port"`
	Count           uint   `json:"count"`
	SocketActivated bool   `json:"socketActivated"`
}

type Isolator struct {
	// Name is the name of the Isolator type as defined in the specification.
	Name string `json:"name"`
	// ValueRaw captures the raw JSON value of an Isolator that was
	// unmarshalled. This field is used for unmarshalling only. It MUST NOT
	// be referenced by external users of the Isolator struct. It is
	// exported only to satisfy Go's unfortunate requirement that fields
	// must be capitalized to be unmarshalled successfully.
	ValueRaw *json.RawMessage `json:"value"`
	// value captures the "true" value of the isolator.
	value IsolatorValue
}

type IsolatorValue interface {
//	UnmarshalJSON(b []byte) error
//	AssertValid() error
}

type Annotation struct {
	Name  string   `json:"name"`
	Value string   `json:"value"`
}

type Dependency struct {
	ImageName string    `json:"imageName"`
	ImageID   *Hash     `json:"imageID,omitempty"`
	Labels    []Label   `json:"labels,omitempty"`
	Size      uint      `json:"size,omitempty"`
}

type Hash struct {
	typ string
	Val string
}

type TemplateDesc struct {
	NameSpace  string
	AciName    string
	ServerName string
	ListenMode string
}

var TemplatePath string = "views/aci/index.html"
var AcipathExist bool = true

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
        NameSpace:  namespace,
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
		        r.Acis[i].AciPath  = acipath
		    }
		}
	} else {
    	r.Acis = append(r.Acis, AciDetail{
			ImageName: imgname,	
			SignPath : signpath,
			AciPath  : acipath,
		    })
	}
	r.PubKeysPath = keyspath

	if err := db.Save(r, key); err != nil {
		return err
	}
	return nil
}