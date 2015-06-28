package backend

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/astaxie/beego/config"
	"github.com/google/google-api-go-client/storage/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	//"github.com/google/google-api-go-client/storage/v1"
)

var (
	projectID   string
	bucket      string
	scope       string
	privateKey  []byte
	clientEmail string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	//Reading config file named conf/runtime.conf for backend
	conf, err := config.NewConfig("ini", gopath+"/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		log.Fatalf("GCS reading conf/runtime.conf err %v", err)
	}

	if projectID = conf.String("Googlecloud::projectid"); projectID == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get projectID err, is nil")
	}

	//Get config var for jsonKeyFile, bucketName, projectID, which should be used later in oauth and get obj
	if bucket = conf.String("Googlecloud::bucket"); bucket == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get bucket err, is nil")
	}

	if scope = conf.String("Googlecloud::scope"); scope == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get privateKey err, is nil")
	}

	var privateKeyFile string
	if privateKeyFile = conf.String("Googlecloud::privatekey"); privateKeyFile == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get privateKey err, is nil")
	}
	privateKey, err = ioutil.ReadFile(gopath + "/src/github.com/containerops/dockyard/conf/" + privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	if clientEmail = conf.String("Googlecloud::clientemail"); clientEmail == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get clientEmail err, is nil")
	}

	g_injector.Bind("googlecloudsave", googlecloudsave)
}

func googlecloudsave(file string) (url string, err error) {

	s := []string{scope}

	conf := jwt.Config{
		Email:      clientEmail,
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
	object := &storage.Object{Name: objectName}

	// Insert an object into a bucket.
	fileDes, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening %q: %v", file, err)
	}
	objs, err := service.Objects.Insert(bucket, object).Media(fileDes).Do()
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
