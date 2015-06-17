package gcs

import (
	"fmt"
	"github.com/astaxie/beego/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	storage "google.golang.org/api/storage/v1"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	jsonKeyFile string
	bucketName  string
	projectID   string
	driverType  string
)

/*
//Should Open this const config to replace below one when use testgcs.go
const (
	jsonfile   = "./gcs/key.json"
	bucketName = "dockyad-example-bucket"
	projectID  = "dockyad-test"
)
*/
func init() {

	//Reading config file named conf/runtime.conf for backend
	conf, err := config.NewConfig("ini", "./runtime.conf")
	if err != nil {
		log.Fatalf("GCS reading conf/runtime.conf err %v", err)
	}

	driverType = conf.String("backenddriver")
	if driverType == "" {
		log.Fatalf("GCS reading conf/runtime.conf, get driverType is nil")
	}
	//Get config var for jsonKeyFile, bucketName, projectID, which should be used later in oauth and get obj
	if jsonKeyFile = conf.String(driverType + "::jsonkeyfile"); jsonKeyFile == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get jsonKeyFile err, is nil")
	}

	if bucketName = conf.String(driverType + "::bucketname"); bucketName == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get bucketName err, is nil")
	}

	if projectID := conf.String(driverType + "::projectid"); projectID == "" {
		log.Fatalf("GCS reading conf/runtime.conf, GCS get projectID err, is nil")
	}
}

func Gcssave(file string) (url string, err error) {

	//read json key(key.json) to do oauth according JWT
	data, err := ioutil.ReadFile(jsonKeyFile)
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		log.Fatal(err)
	}

	//new storage service and token, we dont need context here
	client := conf.Client(oauth2.NoContext)
	gcsToken, err := conf.TokenSource(oauth2.NoContext).Token()
	service, err := storage.New(client)
	if err != nil {
		log.Fatalf("GCS unable to create storage service: %v", err)
	}

	// If the bucket already exists and the user has access,  don't try to create it.
	if _, err := service.Buckets.Get(bucketName).Do(); err != nil {
		// If bucket is not exist, Create a bucket.
		if _, err := service.Buckets.Insert(projectID, &storage.Bucket{Name: bucketName}).Do(); err != nil {
			log.Fatalf("GCS failed creating bucket %s: %v", bucketName, err)
		}
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
	objs, err := service.Objects.Insert(bucketName, object).Media(fileDes).Do()
	if err != nil {
		log.Fatalf("GCS Objects.Insert failed: %v", err)
	}
	retUrl := objs.MediaLink + "&access_token=" + gcsToken.AccessToken
	fmt.Println(fmt.Sprintf("GCS tmpUrl=%s", retUrl))

	if err != nil {
		return "", err
	} else {
		return retUrl, nil
	}
}
