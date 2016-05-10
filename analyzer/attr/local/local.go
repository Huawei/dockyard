package local

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/containerops/dockyard/analyzer/attr"
)

type FileBackend struct {
	file *os.File
}

func NewFileBackend(file *os.File) *FileBackend {
	return &FileBackend{
		file: file,
	}
}

func (lb *FileBackend) GetImageInfo(dockerURL string) ([]string, *attr.ParsedDockerURL, error) {
	parsedDockerURL := attr.ParseDockerURL(dockerURL)

	appImageID, parsedDockerURL, err := getImageID(lb.file, parsedDockerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting ImageID: %v", err)
	}

	ancestry, err := getAncestry(lb.file, appImageID)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting ancestry: %v", err)
	}

	return ancestry, parsedDockerURL, nil
}

func (lb *FileBackend) GetLayerInfo(layerID string, dockerURL *attr.ParsedDockerURL) (*attr.DockerImageData, error) {

	j, err := getJson(lb.file, layerID)
	if err != nil {
		return nil, err
	}

	layerData := attr.DockerImageData{}
	if err := json.Unmarshal(j, &layerData); err != nil {
		return nil, err
	}

	return &layerData, nil
}

func getImageID(file *os.File, dockerURL *attr.ParsedDockerURL) (string, *attr.ParsedDockerURL, error) {
	type tags map[string]string
	type apps map[string]tags

	_, err := file.Seek(0, 0)
	if err != nil {
		return "", nil, fmt.Errorf("error seeking file: %v", err)
	}

	var imageID string
	var appName string
	reposWalker := func(t *TarFile) error {
		if filepath.Clean(t.Name()) == "repositories" {
			repob, err := ioutil.ReadAll(t.TarStream)
			if err != nil {
				return fmt.Errorf("error reading repositories file: %v", err)
			}

			var repositories apps
			if err := json.Unmarshal(repob, &repositories); err != nil {
				return fmt.Errorf("error unmarshaling repositories file")
			}

			if dockerURL == nil {
				n := len(repositories)
				switch {
				case n == 1:
					for key, _ := range repositories {
						appName = key
					}
				case n > 1:
					var appNames []string
					for key, _ := range repositories {
						appNames = append(appNames, key)
					}
					return fmt.Errorf("several images found, use option --image with one of:\n\n%s", strings.Join(appNames, "\n"))
				default:
					return fmt.Errorf("no images found")
				}
			} else {
				appName = dockerURL.ImageName
			}

			tag := "latest"
			if dockerURL != nil {
				tag = dockerURL.Tag
			}

			app, ok := repositories[appName]
			if !ok {
				return fmt.Errorf("app %q not found", appName)
			}

			_, ok = app[tag]
			if !ok {
				if len(app) == 1 {
					for key, _ := range app {
						tag = key
					}
				} else {
					return fmt.Errorf("tag %q not found", tag)
				}
			}

			if dockerURL == nil {
				dockerURL = &attr.ParsedDockerURL{
					IndexURL:  "",
					Tag:       tag,
					ImageName: appName,
				}
			}

			imageID = string(app[tag])
		}

		return nil
	}

	tr := tar.NewReader(file)
	if err := Walk(*tr, reposWalker); err != nil {
		return "", nil, err
	}

	if imageID == "" {
		return "", nil, fmt.Errorf("repositories file not found")
	}

	return imageID, dockerURL, nil
}

func getJson(file *os.File, layerID string) ([]byte, error) {
	jsonPath := path.Join(layerID, "json")
	return getTarFileBytes(file, jsonPath)
}

func getTarFileBytes(file *os.File, path string) ([]byte, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		fmt.Errorf("error seeking file: %v", err)
	}

	var fileBytes []byte
	fileWalker := func(t *TarFile) error {
		if filepath.Clean(t.Name()) == path {
			fileBytes, err = ioutil.ReadAll(t.TarStream)
			if err != nil {
				return err
			}
		}

		return nil
	}

	tr := tar.NewReader(file)
	if err := Walk(*tr, fileWalker); err != nil {
		return nil, err
	}

	if fileBytes == nil {
		return nil, fmt.Errorf("file %s not found", path)
	}

	return fileBytes, nil
}

func getAncestry(file *os.File, imgID string) ([]string, error) {
	var ancestry []string

	curImgID := imgID

	var err error
	for curImgID != "" {
		ancestry = append(ancestry, curImgID)
		curImgID, err = getParent(file, curImgID)
		if err != nil {
			return nil, err
		}
	}

	return ancestry, nil
}

func getParent(file *os.File, imgID string) (string, error) {
	var parent string

	_, err := file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("error seeking file: %v", err)
	}

	jsonPath := filepath.Join(imgID, "json")
	parentWalker := func(t *TarFile) error {
		if filepath.Clean(t.Name()) == jsonPath {
			jsonb, err := ioutil.ReadAll(t.TarStream)
			if err != nil {
				return fmt.Errorf("error reading layer json: %v", err)
			}

			var dockerData attr.DockerImageData
			if err := json.Unmarshal(jsonb, &dockerData); err != nil {
				return fmt.Errorf("error unmarshaling layer data: %v", err)
			}

			parent = dockerData.Parent
		}

		return nil
	}

	tr := tar.NewReader(file)
	if err := Walk(*tr, parentWalker); err != nil {
		return "", err
	}

	return parent, nil
}
