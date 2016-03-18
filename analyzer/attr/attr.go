package attr

import (
	"strings"
	"time"
)

type DockerImg_Attr struct {
	Layer    string `json:"layer"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Author   string `json:"author"`
	Epoch    string `json:"epoch"`
	Comment  string `json:"comment"`
	Parent   string `json:"parent"`
	Checksum string `json:"checksum"`
	App      App    `json:"app"`
}

type ImgAttr struct {
	Type   string           `json:"type"`
	Layers int              `json:"layers"`
	Attrs  []DockerImg_Attr `json:"attrs"`
}

type Exec []string

type Port struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Port     uint   `json:"port"`
}

type ParsedDockerURL struct {
	IndexURL  string
	ImageName string
	Tag       string
}

type App struct {
	Exec             Exec         `json:"exec"`
	User             string       `json:"user"`
	Group            string       `json:"group"`
	WorkingDirectory string       `json:"workingDirectory,omitempty"`
	Environment      Environment  `json:"environment,omitempty"`
	MountPoints      []MountPoint `json:"mountPoints,omitempty"`
	Ports            []Port       `json:"ports,omitempty"`
}

type MountPoint struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

type Environment []EnvironmentVariable

type EnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (e *Environment) Set(name string, value string) {
	for i, env := range *e {
		if env.Name == name {
			(*e)[i] = EnvironmentVariable{
				Name:  name,
				Value: value,
			}
			return
		}
	}
	env := EnvironmentVariable{
		Name:  name,
		Value: value,
	}
	*e = append(*e, env)
}

func AnalyseDockerManifest(layerData DockerImageData, dockerURL *ParsedDockerURL) (*DockerImg_Attr, error) {
	dockerConfig := layerData.Config
	imgAttr := &DockerImg_Attr{}

	imgAttr.Layer = layerData.ID

	url := dockerURL.IndexURL + "/"
	url += dockerURL.ImageName + "-" + layerData.ID
	imgAttr.Name = url

	tag := dockerURL.Tag
	imgAttr.Tag = tag

	imgAttr.Version = layerData.DockerVersion

	imgAttr.Checksum = layerData.Checksum

	if layerData.OS != "" {
		os := layerData.OS
		imgAttr.OS = os
		if layerData.Architecture != "" {
			arch := layerData.Architecture
			imgAttr.Arch = arch
		}
	}

	if layerData.Author != "" {
		author := layerData.Author
		imgAttr.Author = author
	}

	epoch := time.Unix(0, 0)
	if !layerData.Created.Equal(epoch) {
		createdTime := layerData.Created.Format(time.RFC3339)
		imgAttr.Epoch = createdTime
	}

	if layerData.Comment != "" {
		comment := layerData.Comment
		imgAttr.Comment = comment
	}

	if dockerConfig != nil {
		exec := getExecCommand(dockerConfig.Entrypoint, dockerConfig.Cmd)
		if exec != nil {
			user, group := parseDockerUser(dockerConfig.User)
			var env Environment
			for _, v := range dockerConfig.Env {
				parts := strings.SplitN(v, "=", 2)
				env.Set(parts[0], parts[1])
			}
			app := &App{
				Exec:             exec,
				User:             user,
				Group:            group,
				Environment:      env,
				WorkingDirectory: dockerConfig.WorkingDir,
			}
			app.MountPoints, _ = convertVolumesToMPs(dockerConfig.Volumes)
			app.Ports, _ = getPorts(dockerConfig.ExposedPorts, dockerConfig.PortSpecs)

			imgAttr.App = *app
		}
	}

	if layerData.Parent != "" {
		indexPrefix := ""
		indexPrefix = dockerURL.IndexURL + "/"
		parentImageName := indexPrefix + dockerURL.ImageName + "-" + layerData.Parent
		imgAttr.Parent = parentImageName
	}

	return imgAttr, nil
}
