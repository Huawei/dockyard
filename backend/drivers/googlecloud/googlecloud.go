package googlecloud

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/storage/v1"

	"github.com/containerops/dockyard/backend/drivers"
	"github.com/containerops/wrench/setting"
)

func init() {
	drivers.Register("googlecloudsave", InitFunc)
}

func InitFunc() {
	drivers.InjectReflect.Bind("googlecloudsave", googlecloudsave)
}

func googlecloudsave(filepath string) (url string, err error) {

	var file string
	var filedir string
	fileattr := make(map[int]string)
	for i, key := range strings.Split(filepath, ":") {
		fileattr[i] = key
	}
	if len(fileattr) > 1 {
		file = fileattr[1]
		filedir = fileattr[0]
	} else {
		file = fileattr[0]
	}

	privateKey, err := ioutil.ReadFile(setting.PrivateKeyFilePath + setting.PrivateKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	s := []string{setting.Scope}

	conf := jwt.Config{
		Email:      setting.Clientemail,
		PrivateKey: privateKey,
		Scopes:     s,
		TokenURL:   google.JWTTokenURL,
	}

	//new storage service and token, we dont need context here
	client := conf.Client(oauth2.NoContext)
	gcsToken, err := conf.TokenSource(oauth2.NoContext).Token()
	service, err := storage.New(client)
	if err != nil {
		log.Fatalf("GCS unable to create storage service: %v", err)
	}

	//Split filename as a objectName
	var objectName string
	for _, objectName = range strings.Split(file, "/") {
	}
	if len(fileattr) > 1 {
		objectName = filedir + "/" + objectName
	}
	object := &storage.Object{Name: objectName}
	// Insert an object into a bucket.
	fileDes, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening %q: %v", file, err)
	}
	objs, err := service.Objects.Insert(setting.Bucket, object).Media(fileDes).Do()
	if err != nil {
		log.Fatalf("GCS Objects.Insert failed: %v", err)
	}
	retUrl := objs.MediaLink + "&access_token=" + gcsToken.AccessToken

	if err != nil {
		return "", err
	} else {
		return retUrl, nil
	}
}
