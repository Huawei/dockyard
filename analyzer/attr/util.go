package attr

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func Quote(l []string) []string {
	var quoted []string

	for _, s := range l {
		quoted = append(quoted, fmt.Sprintf("%q", s))
	}

	return quoted
}

func In(list []string, el string) bool {
	return IndexOf(list, el) != -1
}

func IndexOf(list []string, el string) int {
	for i, x := range list {
		if el == x {
			return i
		}
	}
	return -1
}

const (
	dockercfgFileName = ".dockercfg"
)

const (
	defaultTag      = "latest"
	defaultIndexURL = "registry-1.docker.io"
)

func ParseDockerURL(arg string) *ParsedDockerURL {
	if arg == "" {
		return nil
	}

	taglessRemote, tag := parseRepositoryTag(arg)
	if tag == "" {
		tag = defaultTag
	}
	indexURL, imageName := SplitReposName(taglessRemote)

	if indexURL == "" && !strings.Contains(imageName, "/") {
		imageName = "library/" + imageName
	}

	if indexURL == "" {
		indexURL = defaultIndexURL
	}

	return &ParsedDockerURL{
		IndexURL:  indexURL,
		ImageName: imageName,
		Tag:       tag,
	}
}

// splitReposName breaks a reposName into an index name and remote name
func SplitReposName(reposName string) (string, string) {
	nameParts := strings.SplitN(reposName, "/", 2)
	var indexName, remoteName string
	if len(nameParts) == 1 || (!strings.Contains(nameParts[0], ".") &&
		!strings.Contains(nameParts[0], ":") && nameParts[0] != "localhost") {
		// This is a Docker Index repos (ex: samalba/hipache or ubuntu)
		// The URL for the index is different depending on the version of the
		// API used to fetch it, so it cannot be inferred here.
		indexName = ""
		remoteName = reposName
	} else {
		indexName = nameParts[0]
		remoteName = nameParts[1]
	}
	return indexName, remoteName
}

// Get a repos name and returns the right reposName + tag
// The tag can be confusing because of a port in a repository name.
//     Ex: localhost.localdomain:5000/samalba/hipache:latest
func parseRepositoryTag(repos string) (string, string) {
	n := strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, ""
}

func decodeDockerAuth(s string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid auth configuration file")
	}
	user := parts[0]
	password := strings.Trim(parts[1], "\x00")
	return user, password, nil
}

func getHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

// GetDockercfgAuth reads a ~/.dockercfg file and returns the username and password
// of the given docker index server.
func GetAuthInfo(indexServer string) (string, string, error) {
	dockerCfgPath := path.Join(getHomeDir(), dockercfgFileName)

	if _, err := os.Stat(dockerCfgPath); os.IsNotExist(err) {
		return "", "", nil
	}

	j, err := ioutil.ReadFile(dockerCfgPath)
	if err != nil {
		return "", "", err
	}

	var dockerAuth map[string]DockerAuthConfig
	if err := json.Unmarshal(j, &dockerAuth); err != nil {
		return "", "", err
	}

	// the official auth uses the full address instead of the hostname
	officialAddress := "https://" + indexServer + "/v1/"
	if c, ok := dockerAuth[officialAddress]; ok {
		return decodeDockerAuth(c.Auth)
	}

	// try the normal case
	if c, ok := dockerAuth[indexServer]; ok {
		return decodeDockerAuth(c.Auth)
	}

	return "", "", nil
}

func parseDockerUser(dockerUser string) (string, string) {
	// if the docker user is empty assume root user and group
	if dockerUser == "" {
		return "0", "0"
	}

	dockerUserParts := strings.Split(dockerUser, ":")

	// when only the user is given, the docker spec says that the default and
	// supplementary groups of the user in /etc/passwd should be applied.
	// Assume root group for now in this case.
	if len(dockerUserParts) < 2 {
		return dockerUserParts[0], "0"
	}

	return dockerUserParts[0], dockerUserParts[1]
}

func getExecCommand(entrypoint []string, cmd []string) Exec {
	var command []string
	if entrypoint == nil && cmd == nil {
		return nil
	}
	command = append(entrypoint, cmd...)
	// non-absolute paths are not allowed, fallback to "/bin/sh -c command"
	if len(command) > 0 && !filepath.IsAbs(command[0]) {
		command_prefix := []string{"/bin/sh", "-c"}
		quoted_command := Quote(command)
		command = append(command_prefix, strings.Join(quoted_command, " "))
	}
	return command
}

func getPorts(dockerExposedPorts map[string]struct{}, dockerPortSpecs []string) ([]Port, error) {
	ports := []Port{}

	for ep := range dockerExposedPorts {
		aPort, err := parseDockerPort(ep)
		if err != nil {
			return nil, err
		}
		ports = append(ports, *aPort)
	}

	if dockerExposedPorts == nil && dockerPortSpecs != nil {
		fmt.Println("warning: docker image uses deprecated PortSpecs field")
		for _, ep := range dockerPortSpecs {
			aPort, err := parseDockerPort(ep)
			if err != nil {
				return nil, err
			}
			ports = append(ports, *aPort)
		}
	}

	return ports, nil
}

func parseDockerPort(dockerPort string) (*Port, error) {
	var portString string
	proto := "tcp"
	sp := strings.Split(dockerPort, "/")
	if len(sp) < 2 {
		portString = dockerPort
	} else {
		proto = sp[1]
		portString = sp[0]
	}

	port, err := strconv.ParseUint(portString, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("error parsing port %q: %v", portString, err)
	}

	sn := strings.ToLower(dockerPort)

	parsedPort := &Port{
		Name:     sn,
		Protocol: proto,
		Port:     uint(port),
	}

	return parsedPort, nil
}

func convertVolumesToMPs(dockerVolumes map[string]struct{}) ([]MountPoint, error) {
	mps := []MountPoint{}
	dup := make(map[string]int)

	for p := range dockerVolumes {
		sn := filepath.Join("volume", p)

		// check for duplicate names
		if i, ok := dup[sn]; ok {
			dup[sn] = i + 1
			sn = fmt.Sprintf("%s-%d", sn, i)
		} else {
			dup[sn] = 1
		}

		mp := MountPoint{
			Name: sn,
			Path: p,
		}

		mps = append(mps, mp)
	}

	return mps, nil
}
