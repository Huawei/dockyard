package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/containerops/dockyard/analyzer/attr"
)

type RepoData struct {
	Tokens    []string
	Endpoints []string
	Cookie    []string
}

func (rb *RepoBackend) getLayerInfoV1(layerID string, dockerURL *attr.ParsedDockerURL) (*attr.DockerImageData, error) {

	j, _, err := rb.getJsonV1(layerID, rb.repoData.Endpoints[0], rb.repoData)
	if err != nil {
		return nil, nil
	}

	layerData := attr.DockerImageData{}
	if err := json.Unmarshal(j, &layerData); err != nil {
		return nil, nil
	}

	return &layerData, nil
}

func (rb *RepoBackend) getImageInfoV1(dockerURL *attr.ParsedDockerURL) ([]string, *attr.ParsedDockerURL, error) {
	repoData, err := rb.getRepoDataV1(dockerURL.IndexURL, dockerURL.ImageName)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting repository data: %v", err)
	}

	// TODO(iaguis) check more endpoints
	appImageID, err := rb.getImageIDFromTagV1(repoData.Endpoints[0], dockerURL.ImageName, dockerURL.Tag, repoData)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting ImageID from tag %s: %v", dockerURL.Tag, err)
	}

	ancestry, err := rb.getAncestryV1(appImageID, repoData.Endpoints[0], repoData)
	if err != nil {
		return nil, nil, err
	}

	rb.repoData = repoData

	return ancestry, dockerURL, nil
}

func (rb *RepoBackend) getRepoDataV1(indexURL string, remote string) (*RepoData, error) {
	client := &http.Client{}
	repoURL := rb.protocol() + path.Join(indexURL, "v1", "repositories", remote, "images")

	req, err := http.NewRequest("GET", repoURL, nil)
	if err != nil {
		return nil, err
	}

	if rb.username != "" && rb.password != "" {
		req.SetBasicAuth(rb.username, rb.password)
	}

	req.Header.Set("X-Docker-Token", "true")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP code: %d, URL: %s", res.StatusCode, req.URL)
	}

	var tokens []string
	if res.Header.Get("X-Docker-Token") != "" {
		tokens = res.Header["X-Docker-Token"]
	}

	var cookies []string
	if res.Header.Get("Set-Cookie") != "" {
		cookies = res.Header["Set-Cookie"]
	}

	var endpoints []string
	if res.Header.Get("X-Docker-Endpoints") != "" {
		endpoints = makeEndpointsListV1(res.Header["X-Docker-Endpoints"])
	} else {
		// Assume same endpoint
		endpoints = append(endpoints, indexURL)
	}

	return &RepoData{
		Endpoints: endpoints,
		Tokens:    tokens,
		Cookie:    cookies,
	}, nil
}

func (rb *RepoBackend) getImageIDFromTagV1(registry string, appName string, tag string, repoData *RepoData) (string, error) {
	client := &http.Client{}
	// we get all the tags instead of directly getting the imageID of the
	// requested one (.../tags/TAG) because even though it's specified in the
	// Docker API, some registries (e.g. Google Container Registry) don't
	// implement it.
	req, err := http.NewRequest("GET", rb.protocol()+path.Join(registry, "repositories", appName, "tags"), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get Image ID: %s, URL: %s", err, req.URL)
	}

	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get Image ID: %s, URL: %s", err, req.URL)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("HTTP code: %d. URL: %s", res.StatusCode, req.URL)
	}

	j, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	var tags map[string]string

	if err := json.Unmarshal(j, &tags); err != nil {
		return "", fmt.Errorf("error unmarshaling: %v", err)
	}

	imageID, ok := tags[tag]
	if !ok {
		return "", fmt.Errorf("tag %s not found", tag)
	}

	return imageID, nil
}

func (rb *RepoBackend) getAncestryV1(imgID, registry string, repoData *RepoData) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rb.protocol()+path.Join(registry, "images", imgID, "ancestry"), nil)
	if err != nil {
		return nil, err
	}

	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP code: %d. URL: %s", res.StatusCode, req.URL)
	}

	var ancestry []string

	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read downloaded json: %s (%s)", err, j)
	}

	if err := json.Unmarshal(j, &ancestry); err != nil {
		return nil, fmt.Errorf("error unmarshaling: %v", err)
	}

	return ancestry, nil
}

func (rb *RepoBackend) getJsonV1(imgID, registry string, repoData *RepoData) ([]byte, int64, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", rb.protocol()+path.Join(registry, "images", imgID, "json"), nil)
	if err != nil {
		return nil, -1, err
	}
	setAuthTokenV1(req, repoData.Tokens)
	setCookieV1(req, repoData.Cookie)
	res, err := client.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, -1, fmt.Errorf("HTTP code: %d, URL: %s", res.StatusCode, req.URL)
	}

	imageSize := int64(-1)

	if hdr := res.Header.Get("X-Docker-Size"); hdr != "" {
		imageSize, err = strconv.ParseInt(hdr, 10, 64)
		if err != nil {
			return nil, -1, err
		}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to read downloaded json: %v (%s)", err, b)
	}

	return b, imageSize, nil
}

func setAuthTokenV1(req *http.Request, token []string) {
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Token "+strings.Join(token, ","))
	}
}

func setCookieV1(req *http.Request, cookie []string) {
	if req.Header.Get("Cookie") == "" {
		req.Header.Set("Cookie", strings.Join(cookie, ""))
	}
}

func makeEndpointsListV1(headers []string) []string {
	var endpoints []string

	for _, ep := range headers {
		endpointsList := strings.Split(ep, ",")
		for _, endpointEl := range endpointsList {
			endpoints = append(
				endpoints,
				path.Join(strings.TrimSpace(endpointEl), "v1"))
		}
	}

	return endpoints
}
