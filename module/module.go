package module

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/openpgp"

	"github.com/containerops/dockyard/models"
	"github.com/containerops/wrench/setting"
	"github.com/containerops/wrench/utils"
)

func ParseManifest(data []byte, namespace, repository, tag string) (error, int) {
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err, 0
	}

	schemaVersion := int(manifest["schemaVersion"].(float64))
	if schemaVersion == 1 {
		for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
			v := manifest["history"].([]interface{})[k]
			compatibility := v.(map[string]interface{})["v1Compatibility"].(string)

			var image map[string]interface{}
			if err := json.Unmarshal([]byte(compatibility), &image); err != nil {
				return err, 0
			}

			i := map[string]string{}
			r := new(models.Repository)

			if k == 0 {
				i["Tag"] = tag
			}
			i["id"] = image["id"].(string)

			if err := r.PutJSONFromManifests(i, namespace, repository); err != nil {
				return err, 0
			}

			if k == 0 {
				if err := r.PutTagFromManifests(image["id"].(string), namespace, repository, tag, string(data), schemaVersion); err != nil {
					return err, 0
				}
			}
		}
	} else if schemaVersion == 2 {
		r := new(models.Repository)
		if err := r.PutTagFromManifests("schemaV2", namespace, repository, tag, string(data), schemaVersion); err != nil {
			return err, 0
		}
	} else {
		return fmt.Errorf("Invalid schema version"), 0
	}

	return nil, schemaVersion
}

func CopyImgLayer(srcPath, srcFile, dstPath, dstFile string, reqbody []byte) (int, error) {
	if !utils.IsDirExist(dstPath) {
		os.MkdirAll(dstPath, os.ModePerm)
	}

	if utils.IsFileExist(dstFile) {
		os.Remove(dstFile)
	}

	var data []byte
	if _, err := os.Stat(srcFile); err == nil {
		data, _ = ioutil.ReadFile(srcFile)
		if err := ioutil.WriteFile(dstFile, data, 0777); err != nil {
			return 0, err
		}
		os.RemoveAll(srcPath)
	} else {
		data = reqbody
		if err := ioutil.WriteFile(dstFile, data, 0777); err != nil {
			return 0, err
		}
	}

	return len(data), nil
}

//codes as below are ported to support for docker to parse request URL,and it would be update soon
func parseIP(ipStr string) net.IP {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		fmt.Errorf("invalid remote IP address: %q", ipStr)
	}
	return ip
}

func RemoteAddr(r *http.Request) string {
	if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
		proxies := strings.Split(prior, ",")
		if len(proxies) > 0 {
			remoteAddr := strings.Trim(proxies[0], " ")
			if parseIP(remoteAddr) != nil {
				return remoteAddr
			}
		}
	}

	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		if parseIP(realIP) != nil {
			return realIP
		}
	}

	return r.RemoteAddr
}

const (
	RouteNameBase            = "base"
	RouteNameBlob            = "blob"
	RouteNameManifest        = "manifest"
	RouteNameTags            = "tags"
	RouteNameBlobUpload      = "blob-upload"
	RouteNameBlobUploadChunk = "blob-upload-chunk"
)

type URLBuilder struct {
	root   *url.URL
	router *mux.Router
}

type RouteDescriptor struct {
	Name string
	Path string
}

var RepositoryNameComponentRegexp = regexp.MustCompile(`[a-z0-9]+(?:[._-][a-z0-9]+)*`)
var RepositoryNameRegexp = regexp.MustCompile(`(?:` + RepositoryNameComponentRegexp.String() + `/)*` + RepositoryNameComponentRegexp.String())
var TagNameRegexp = regexp.MustCompile(`[\w][\w.-]{0,127}`)
var DigestRegexp = regexp.MustCompile(`[a-zA-Z0-9-_+.]+:[a-fA-F0-9]+`)

var routeDescriptors = []RouteDescriptor{
	{
		Name: RouteNameBase,
		Path: "/v2/",
	},
	{
		Name: RouteNameBlob,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/{digest:" + DigestRegexp.String() + "}",
	},
	{
		Name: RouteNameManifest,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/manifests/{reference:" + TagNameRegexp.String() + "|" + DigestRegexp.String() + "}",
	},
	{
		Name: RouteNameTags,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/tags/list",
	},
	{
		Name: RouteNameBlobUpload,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/uploads/",
	},
	{
		Name: RouteNameBlobUploadChunk,
		Path: "/v2/{name:" + RepositoryNameRegexp.String() + "}/blobs/uploads/{uuid:[a-zA-Z0-9-_.=]+}",
	},
}

func NewURLBuilderFromRequest(r *http.Request) *URLBuilder {
	var scheme string

	forwardedProto := r.Header.Get("X-Forwarded-Proto")
	switch {
	case len(forwardedProto) > 0:
		scheme = forwardedProto
	case r.TLS != nil:
		scheme = "https"
	case len(r.URL.Scheme) > 0:
		scheme = r.URL.Scheme
	default:
		scheme = "http"
	}

	host := r.Host
	forwardedHost := r.Header.Get("X-Forwarded-Host")
	if len(forwardedHost) > 0 {
		hosts := strings.SplitN(forwardedHost, ",", 2)
		host = strings.TrimSpace(hosts[0])
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   host,
	}
	/*
		basePath := routeDescriptorsMap[RouteNameBase].Path
		requestPath := r.URL.Path
		index := strings.Index(requestPath, basePath)
		if index > 0 {
			u.Path = requestPath[0 : index+1]
		}
	*/
	return NewURLBuilder(u)
}

func Router() *mux.Router {
	return RouterWithPrefix("")
}

func RouterWithPrefix(prefix string) *mux.Router {
	rootRouter := mux.NewRouter()
	router := rootRouter
	if prefix != "" {
		router = router.PathPrefix(prefix).Subrouter()
	}

	router.StrictSlash(true)

	for _, descriptor := range routeDescriptors {
		router.Path(descriptor.Path).Name(descriptor.Name)
	}

	return rootRouter
}

func NewURLBuilder(root *url.URL) *URLBuilder {
	return &URLBuilder{
		root:   root,
		router: Router(),
	}
}

func (ub *URLBuilder) BuildBlobURL(name string, dgst string) (string, error) {
	route := ub.cloneRoute(RouteNameBlob)

	layerURL, err := route.URL("name", name, "digest", dgst)
	if err != nil {
		return "", err
	}

	return layerURL.String(), nil
}

func (ub *URLBuilder) BuildManifestURL(name, reference string) (string, error) {
	route := ub.cloneRoute(RouteNameManifest)

	manifestURL, err := route.URL("name", name, "reference", reference)
	if err != nil {
		return "", err
	}

	return manifestURL.String(), nil
}

func (ub *URLBuilder) cloneRoute(name string) clonedRoute {
	route := new(mux.Route)
	root := new(url.URL)

	*route = *ub.router.GetRoute(name)
	*root = *ub.root

	return clonedRoute{Route: route, root: root}
}

type clonedRoute struct {
	*mux.Route
	root *url.URL
}

func (cr clonedRoute) URL(pairs ...string) (*url.URL, error) {
	routeURL, err := cr.Route.URL(pairs...)
	if err != nil {
		return nil, err
	}

	if routeURL.Scheme == "" && routeURL.User == nil && routeURL.Host == "" {
		routeURL.Path = routeURL.Path[1:]
	}

	return cr.root.ResolveReference(routeURL), nil
}

//codes as below are implemented to support ACI storage
func VerifyAciSignature(acipath, signpath, pubkeyspath string) error {
	files, err := ioutil.ReadDir(pubkeyspath)
	if err != nil {
		return fmt.Errorf("Read pubkeys directory failed: %v", err.Error())
	}

	if len(files) <= 0 {
		return fmt.Errorf("No pubkey file found in %v", pubkeyspath)
	}

	var keyring openpgp.EntityList
	for _, file := range files {
		pubkeyfile, err := os.Open(pubkeyspath + "/" + file.Name())
		if err != nil {
			return err
		}
		defer pubkeyfile.Close()

		keyList, err := openpgp.ReadArmoredKeyRing(pubkeyfile)
		if err != nil {
			return err
		}

		if len(keyList) < 1 {
			return fmt.Errorf("Missing opengpg entity")
		}

		keyring = append(keyring, keyList[0])
	}

	acifile, err := os.Open(acipath)
	if err != nil {
		return fmt.Errorf("Open ACI file failed: %v", err.Error())
	}
	defer acifile.Close()

	signfile, err := os.Open(signpath)
	if err != nil {
		return fmt.Errorf("Open signature file failed: %v", err.Error())
	}
	defer signfile.Close()

	if _, err := acifile.Seek(0, 0); err != nil {
		return fmt.Errorf("Seek ACI file failed: %v", err)
	}
	if _, err := signfile.Seek(0, 0); err != nil {
		return fmt.Errorf("Seek signature file: %v", err)
	}

	//Verify detached signature which default is ASCII format
	_, err = openpgp.CheckArmoredDetachedSignature(keyring, acifile, signfile)
	if err == io.EOF {
		if _, err := acifile.Seek(0, 0); err != nil {
			return fmt.Errorf("Seek ACI file failed: %v", err)
		}
		if _, err := signfile.Seek(0, 0); err != nil {
			return fmt.Errorf("Seek signature file: %v", err)
		}

		//try to verify detached signature with binary format
		_, err = openpgp.CheckDetachedSignature(keyring, acifile, signfile)
	}
	if err == io.EOF {
		return fmt.Errorf("Signature format is invalid")
	}

	return err
}

func CheckClientStatus(reqbody []byte) error {
	clientmsg := new(models.CompleteMsg)
	if err := json.Unmarshal(reqbody, &clientmsg); err != nil {
		return fmt.Errorf("%v", err.Error())
	}

	if !clientmsg.Success {
		return fmt.Errorf("%v", clientmsg.Reason)
	}

	return nil
}

func FillRespMsg(result bool, clientreason, serverreason string) models.CompleteMsg {
	msg := models.CompleteMsg{
		Success:      result,
		Reason:       clientreason,
		ServerReason: serverreason,
	}
	return msg
}

func CleanCache(imageId string) {
	acipath := fmt.Sprintf("%v/acis/%v", setting.ImagePath, imageId)
	os.RemoveAll(acipath)
}

func GetPubkeysPath(namespace, repository string) string {
	return fmt.Sprintf("%v/acis/pubkeys/%v/%v", setting.ImagePath, namespace, repository)
}

func GetImagePath(imageId string) string {
	return fmt.Sprintf("%v/acis/%v", setting.ImagePath, imageId)
}

func GetManifestPath(imageId string) string {
	return fmt.Sprintf("%v/acis/%v/manifest", setting.ImagePath, imageId)
}

func GetSignaturePath(imageId, signfile string) string {
	return fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, imageId, signfile)
}

func GetAciPath(imageId, acifile string) string {
	return fmt.Sprintf("%v/acis/%v/%v", setting.ImagePath, imageId, acifile)
}
