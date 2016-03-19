package redis

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/redis.v3"

	"github.com/containerops/dockyard/utils/db/factory"
)

var (
	Client *redis.Client
)

type redisdrv struct{}

func init() {
	factory.Register("redis", &redisdrv{})
}

func Key(obj interface{}) (result string) {
	object := reflect.TypeOf(obj).Elem().Name()
	s := reflect.ValueOf(obj).Elem()
	typeOfS := s.Type()
	keys := []string{}

	switch strings.ToLower(object) {
	case "repository":
		for k := 0; k < s.NumField(); k++ {
			t := typeOfS.Field(k).Name
			if t == "Namespace" || t == "Repository" {
				keys = append(keys, s.Field(k).Interface().(string))
			}
		}
		result = fmt.Sprintf("REPO-%s-%s", keys[0], keys[1])
	case "image":
		for k := 0; k < s.NumField(); k++ {
			t := typeOfS.Field(k).Name
			if t == "ImageId" {
				keys = append(keys, s.Field(k).Interface().(string))
			}
		}
		result = fmt.Sprintf("IMAGE-%s", keys[0])
	case "tag":
		for k := 0; k < s.NumField(); k++ {
			t := typeOfS.Field(k).Name
			if t == "Namespace" || t == "Repository" || t == "Tag" {
				keys = append(keys, s.Field(k).Interface().(string))
			}
		}
		result = fmt.Sprintf("TAG-%s-%s-%s", keys[0], keys[1], keys[2])
	default:
		result = ""
	}

	return result
}

func (r *redisdrv) RegisterModel(models ...interface{}) {
	return
}

func (r *redisdrv) InitDB(driver, user, passwd, uri, name string, partition int64) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: passwd,
		DB:       partition,
	})

	if _, err := Client.Ping().Result(); err != nil {
		return err
	} else {
		return nil
	}
}

func (r *redisdrv) Get(obj interface{}, params ...string) (bool, error) {
	key := Key(obj)
	result, err := Client.Get(key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		} else {
			return false, err
		}
	}

	if err = json.Unmarshal([]byte(result), &obj); err != nil {
		return false, err
	}

	return true, nil
}

func (r *redisdrv) Save(obj interface{}, params ...string) error {
	result, err := json.Marshal(&obj)
	if err != nil {
		return err
	}

	key := Key(obj)
	if _, err := Client.Set(key, string(result), 0).Result(); err != nil {
		return err
	}

	return nil
}

func (r *redisdrv) Insert(obj interface{}) error {
	return r.Save(obj)
}

func (r *redisdrv) Update(obj interface{}, params ...string) error {
	return r.Save(obj, params...)
}
