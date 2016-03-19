package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type NotificationsCtx struct {
	Name      string         `json:"name,omitempty"`
	Endpoints []EndpointDesc `json:"endpoints,omitempty"`
}

type EndpointDesc struct {
	Name      string        `json:"name"`
	URL       string        `json:"url"`
	Headers   http.Header   `json:"headers"`
	Timeout   time.Duration `json:"timeout"`
	Threshold int           `json:"threshold"`
	Backoff   time.Duration `json:"backoff"`
	EventDB   string        `json:"eventdb"`
	Disabled  bool          `json:"disabled"`
}

type AuthorDesc map[string]interface{}

type AuthorsCtx map[string]AuthorDesc

type Desc struct {
	Notifications NotificationsCtx `json:"notifications,omitempty"`
	Authors       AuthorsCtx       `json:"auth,omitempty"`
}

var JSONConfCtx Desc

func GetConfFromJSON(path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	if err := json.Unmarshal(buf, &JSONConfCtx); err != nil {
		return fmt.Errorf("err: %v", err.Error())
	}

	return nil
}

func (auth AuthorsCtx) Name() (name string) {
	name = ""
	for key, _ := range auth {
		name = key
		break
	}
	return
}
