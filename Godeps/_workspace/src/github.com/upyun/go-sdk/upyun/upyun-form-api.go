package upyun

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// UPYUN HTTP FORM API Client
type UpYunForm struct {
	// Core
	upYunHTTPCore

	Secret string
	Bucket string
}

// Response from UPYUN Form API Server
type FormAPIResp struct {
	Code      int    `json:"code"`
	Msg       string `json:"message"`
	Url       string `json:"url"`
	Timestamp int64  `json:"time"`
	ImgWidth  int    `json:"image-width"`
	ImgHeight int    `json:"image-height"`
	ImgFrames int    `json:"image-frames"`
	ImgType   string `json:"image-type"`
	Sign      string `json:"sign"`
}

// NewUpYunForm return a UPYUN Form API client given
// a form api key and bucket name. As Default, endpoint
// is set to Auto, http client connection timeout is
// set to defalutConnectionTimeout which is equal to
// 60 seconds.
func NewUpYunForm(bucket, secret string) *UpYunForm {
	upm := &UpYunForm{
		Secret: secret,
		Bucket: bucket,
	}

	upm.httpClient = &http.Client{}
	upm.SetTimeout(defaultConnectTimeout)
	upm.SetEndpoint(Auto)

	return upm
}

// SetEndpoint sets the request endpoint to UPYUN Form API Server.
func (u *UpYunForm) SetEndpoint(ed int) error {
	if ed >= Auto && ed <= Ctt {
		u.endpoint = fmt.Sprintf("v%d.api.upyun.com", ed)
		return nil
	}

	return errors.New("Invalid endpoint, pick from Auto, Telecom, Cnc, Ctt")
}

// Put posts a http form request given reader, save path,
// expiration, other options and returns a FormAPIResp pointer.
func (uf *UpYunForm) Put(fpath, saveas string, expireAfter int64,
	options map[string]string) (*FormAPIResp, error) {
	if options == nil {
		options = make(map[string]string)
	}

	options["bucket"] = uf.Bucket
	options["save-key"] = saveas
	options["expiration"] = strconv.FormatInt(time.Now().Unix()+expireAfter, 10)

	args, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	policy := base64.StdEncoding.EncodeToString(args)
	sig := md5Str(policy + "&" + uf.Secret)

	fd, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	url := fmt.Sprintf("http://%s/%s", uf.endpoint, uf.Bucket)
	resp, err := uf.doFormRequest(url, policy, sig, fpath, fd)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode/100 == 2 {
		var formResp FormAPIResp
		if err := json.Unmarshal(buf, &formResp); err != nil {
			return nil, err
		}
		return &formResp, nil
	}

	return nil, errors.New(string(buf))
}
