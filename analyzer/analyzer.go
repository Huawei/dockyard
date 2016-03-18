package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containerops/dockyard/analyzer/attr"
	"github.com/containerops/dockyard/analyzer/attr/local"
	"github.com/containerops/dockyard/analyzer/attr/registry"
)

var AnalyzerChan ShareChannel

func InitAnalyzer() error {

	AnalyzerChan = *NewShareChannel()
	AnalyzerChan.Open()

	return nil
}

type AnalyseBackend interface {
	GetImageInfo(dockerUrl string) ([]string, *attr.ParsedDockerURL, error)
	GetLayerInfo(layerID string, dockerURL *attr.ParsedDockerURL) (*attr.DockerImageData, error)
}

func AnalyseRegistry(dockerURL string, username string, password string, insecure bool) (*attr.ImgAttr, error) {
	repoBackend := registry.NewRepoBackend(username, password, insecure)
	return analyseReal(repoBackend, dockerURL)
}

func AnalyseLocal(dockerURL string, filePath string) (*attr.ImgAttr, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	fileBackend := local.NewFileBackend(f)
	return analyseReal(fileBackend, dockerURL)
}

func AnalyseManifestFile(jsonIn string) (*attr.ImgAttr, error) {
	imgData := attr.DockerImageData{}
	err := json.Unmarshal([]byte(jsonIn), &imgData)
	if nil != err {
		return nil, err
	}

	var attrs []attr.DockerImg_Attr

	parsedDockerURL := &attr.ParsedDockerURL{}

	imgLayerAttr, _ := attr.AnalyseDockerManifest(imgData, parsedDockerURL)

	attrs = append(attrs, *imgLayerAttr)

	imgAttr := &attr.ImgAttr{Type: "Docker Image", Layers: 1, Attrs: attrs}

	return imgAttr, nil
}

// GetIndexName returns the docker index server from a docker URL.
func GetIndexName(dockerURL string) string {
	index, _ := attr.SplitReposName(dockerURL)
	return index
}

// GetDockercfgAuth reads a ~/.dockercfg file and returns the username and password
// of the given docker index server.
func GetDockercfgAuth(indexServer string) (string, string, error) {
	return attr.GetAuthInfo(indexServer)
}

func analyseReal(backend AnalyseBackend, dockerURL string) (*attr.ImgAttr, error) {
	ancestry, parsedDockerURL, err := backend.GetImageInfo(dockerURL)
	if err != nil {
		return nil, err
	}

	var attrs []attr.DockerImg_Attr

	for i := len(ancestry) - 1; i >= 0; i-- {
		layerID := ancestry[i]

		layerData, _ := backend.GetLayerInfo(layerID, parsedDockerURL)

		imgLayerAttr, _ := attr.AnalyseDockerManifest(*layerData, parsedDockerURL)

		attrs = append(attrs, *imgLayerAttr)
	}

	imgAttr := &attr.ImgAttr{Type: "Docker Image", Layers: len(ancestry), Attrs: attrs}

	return imgAttr, nil
}

// striplayerID strips the layer ID from an app name:
//
// myregistry.com/organization/app-name-85738f8f9a7f1b04b5329c590ebcb9e425925c6d0984089c43a022de4f19c281
// myregistry.com/organization/app-name
func stripLayerID(layerName string) string {
	n := strings.LastIndex(layerName, "-")
	return layerName[:n]
}
