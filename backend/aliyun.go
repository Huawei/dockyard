package backend

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/config"
)

var (
	g_aliEndpoint        string
	g_aliBucket          string
	g_aliAccessKeyID     string
	g_aliAccessKeySecret string
)

func init() {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Errorf("read env GOPATH fail")
		os.Exit(1)
	}
	err := aligetconfig(gopath + "/src/github.com/containerops/dockyard/conf/runtime.conf")
	if err != nil {
		fmt.Errorf("read config file conf/runtime.conf fail:" + err.Error())
		os.Exit(1)
	}
	g_injector.Bind("alicloudsave", alicloudsave)
}

func aligetconfig(conffile string) (err error) {
	conf, err := config.NewConfig("ini", conffile)
	if err != nil {
		return err
	}

	g_aliEndpoint = conf.String("alicloud::endpoint")
	if g_aliEndpoint == "" {
		return errors.New("read config file's endpoint failed!")
	}

	g_aliBucket = conf.String("alicloud::bucket")
	if g_aliBucket == "" {
		return errors.New("read config file's bucket failed!")
	}

	g_aliAccessKeyID = conf.String("alicloud::accessKeyID")
	if g_aliAccessKeyID == "" {
		return errors.New("read config file's accessKeyID failed!")
	}

	g_aliAccessKeySecret = conf.String("alicloud::accessKeysecret")
	if g_aliAccessKeySecret == "" {
		return errors.New("read config file's accessKeysecret failed!")
	}
	return nil
}

func alicloudsave(file string) (url string, err error) {

	client := NewClient(g_aliAccessKeyID, g_aliAccessKeySecret)

	bucket := NewBucket(g_aliBucket, g_aliEndpoint, client)

	var key string
	//get the filename from the file , eg,get "1.txt" from /home/liugenping/1.txt
	for _, key = range strings.Split(file, "/") {

	}
	opath := "/" + g_aliBucket + "/" + key
	url = "http://" + g_aliEndpoint + opath

	headers := map[string]string{}

	err = bucket.PutFile(key, file, headers)

	if nil != err {
		return "", err
	} else {
		return url, nil
	}
}

var resourceQSWhitelist []string = []string{
	"acl",
	"group",
	"uploadId",
	"partNumber",
	"uploads",
	"logging",
	"response-content-type",
	"response-content-language",
	"response-expires",
	"reponse-cache-control",
	"response-content-disposition",
	"response-content-encoding",
}

// Holds OSS client informations
type Client struct {
	accessKeyId     string
	accessKeySecret string
	*http.Client
}

// Holds OSS bucket informations
type Bucket struct {
	name   string
	region string
	*Client
}

// Initialize a new client and sets access_key_id and access_key_secret.
func NewClient(accessKeyId, accessKeySecret string) *Client {
	return &Client{accessKeyId, accessKeySecret, new(http.Client)}
}

// Initialize a new OSS bucket with the given `name`, `region` and `*Client`.
func NewBucket(name string, region string, client *Client) *Bucket {
	return &Bucket{name, region, client}
}

// PUT the given `content` as `object`.
func (b *Bucket) Put(object string, content io.Reader, headers map[string]string) error {
	buffer := new(bytes.Buffer)
	io.Copy(buffer, content)

	header := make(http.Header)
	header.Set("Content-Type", http.DetectContentType(buffer.Bytes()))
	header.Set("Content-Length", strconv.Itoa(buffer.Len()))

	for key, val := range headers {
		header.Set(key, val)
	}

	resp, err := b.do("PUT", b.name, string(b.region), object, header, buffer)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
		return err
	}

	return nil
}

// PUT the file at `filepath` to `object`.
func (b *Bucket) PutFile(object, filepath string, headers map[string]string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return b.Put(object, file, headers)
}

func (c *Client) do(method, bucket, region, object string, header http.Header, body io.Reader) (*http.Response, error) {
	object = strings.Trim(object, "/")
	req, err := http.NewRequest(method, fmt.Sprintf("http://%s.%s/%s", bucket, region, object), body)
	if err != nil {
		return nil, err
	}

	if header == nil {
		header = make(http.Header)
	}
	header.Set("Date", time.Now().UTC().Format(http.TimeFormat))

	resource := fmt.Sprintf("/%s/%s", bucket, object)
	header.Set("Authorization", c.authorization(method, header, resource))

	req.Header = header

	return c.Do(req)
}

// Return an "Authorization" header value in the form of "OSS " + Access Key Id + ":" + Signature
//
// Signature:
//
//     base64(hmac-sha1(Access Key Secret + "\n"
//       + VERB + "\n"
//       + CONTENT-MD5 + "\n"
//       + CONTENT-TYPE + "\n"
//       + DATE + "\n"
//       + CanonicalizedossHeaders
//       + CanonicalizedResource))
func (c *Client) authorization(verb string, header http.Header, resource string) string {
	params := []string{
		verb,
		header.Get("Content-MD5"),
		header.Get("Content-Type"),
		header.Get("Date"),
	}

	signatureStr := strings.Join(params, "\n") + "\n"

	canonicalizedHeaders := c.canonicalizeHeaders(header)
	canonicalizedResource := c.canonicalizeResource(resource)

	if canonicalizedHeaders != "" {
		signatureStr += canonicalizedHeaders
	}
	signatureStr += canonicalizedResource

	h := hmac.New(sha1.New, []byte(c.accessKeySecret))
	h.Write([]byte(signatureStr))

	signedStr := strings.TrimSpace(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	return "OSS " + c.accessKeyId + ":" + signedStr
}

// Generate `CanonicalizedossHeaders`
//
// Spec:
//     - ignore none x-oss- headers
//     - lowercase fields
//     - sort lexicographically
//     - trim whitespace between field and value
//     - join with newline
func (c *Client) canonicalizeHeaders(header http.Header) string {
	ossHeaders := []string{}
	canonicalizedHeaders := ""

	for k, _ := range header {
		field := strings.ToLower(k)

		if strings.HasPrefix(field, "x-oss-") {
			ossHeaders = append(ossHeaders, field)
		}
	}

	sort.Strings(ossHeaders)

	for _, k := range ossHeaders {
		canonicalizedHeaders += k + ":" + header.Get(k) + "\n"
	}

	return canonicalizedHeaders
}

// Generate `CanonicalizedResource`
//
// Spec:
//     - ignore non sub-resource
//     - ignore non override headers
//     - sort lexicographically
func (c *Client) canonicalizeResource(resource string) string {
	u, _ := url.Parse(resource)

	queryies := u.Query()
	query := url.Values{}

	sort.Strings(resourceQSWhitelist)
	for _, q := range resourceQSWhitelist {
		val := queryies.Get(q)
		if val != "" {
			query.Add(q, val)
		}
	}

	u.RawQuery = query.Encode()

	return u.String()
}
