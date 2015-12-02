package db

import (
	"encoding/json"
	"fmt"

	"gopkg.in/redis.v3"
)

const (
	//Dockyard Data Index
	GLOBAL_REPOSITORY_INDEX = "GLOBAL_REPOSITORY_INDEX"
	GLOBAL_IMAGE_INDEX      = "GLOBAL_IMAGE_INDEX"
	GLOBAL_TARSUM_INDEX     = "GLOBAL_TARSUM_INDEX"
	GLOBAL_TAG_INDEX        = "GLOBAL_TAG_INDEX"
	GLOBAL_COMPOSE_INDEX    = "GLOBAL_COMPOSE_INDEX"
	//Sail Data Index
	GLOBAL_USER_INDEX         = "GLOBAL_USER_INDEX"
	GLOBAL_ORGANIZATION_INDEX = "GLOBAL_ORGANIZATION_INDEX"
	GLOBAL_TEAM_INDEX         = "GLOBAL_TEAM_INDEX"
	//Wharf Data Index
	GLOBAL_ADMIN_INDEX = "GLOBAL_ADMIN_INDEX"
	GLOBAL_LOG_INDEX   = "GLOBAL_LOG_INDEX"
)

/*
  [user] : USER-(username)
	[organization] : ORG-(org)
	[team] : TEAM-(org)-(team)
	[repository] : REPO-(namespace)-(repo)
	[image] : IMAGE-(imageId)
	[tag] : TAG-(namespace)-(repo)-(tag)
	[compose] : COMPOSE-(namespace)-(compose)
	[admin] : ADMIN-(username)
	[log] : LOG-(object)
	[lock] : LOCK-(object)
*/

var (
	Client *redis.Client
)

func Key(object string, keys ...string) (result string) {
	switch object {
	case "USER":
	case "user":
		result = fmt.Sprintf("USER-%s", keys[0])
	case "ORG":
	case "ORGANIZATION":
	case "org":
	case "organization":
		result = fmt.Sprintf("ORG-%s", keys[0])
	case "TEAM":
	case "team":
		result = fmt.Sprintf("ORG-%s-%s", keys[0], keys[1])
	case "REPO":
	case "REPOSITORY":
	case "repo":
	case "repository":
		result = fmt.Sprintf("REPO-%s-%s", keys[0], keys[1])
	case "IMAGE":
	case "image":
		result = fmt.Sprintf("IMAGE-%s", keys[0])
	case "TARSUM":
	case "tarsum":
		result = fmt.Sprintf("TARSUM-%s", keys[0])
	case "TAG":
	case "tag":
		result = fmt.Sprintf("TAG-%s-%s-%s", keys[0], keys[1], keys[2])
	case "COMPOSE":
	case "compose":
		result = fmt.Sprintf("COMPOSE-%s-%s", keys[0], keys[1])
	case "ADMIN":
	case "admin":
		result = fmt.Sprintf("ADMIN-%s", keys[0])
	case "LOG":
	case "log":
		result = fmt.Sprintf("LOG-%s", keys[0])
	case "LOCK":
	case "lock":
		result = fmt.Sprintf("LOCK-%s", keys[0])
	default:
		result = ""
	}

	return result
}

func InitDB(addr, passwd string, db int64) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := Client.Ping().Result(); err != nil {
		return err
	} else {
		return nil
	}
}

func Save(obj interface{}, key string) (err error) {
	result, err := json.Marshal(&obj)

	if err != nil {
		return err
	}

	if _, err := Client.Set(key, string(result), 0).Result(); err != nil {
		return err
	}

	return nil
}

func Get(obj interface{}, key string) (err error) {
	result, err := Client.Get(key).Result()

	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(result), &obj); err != nil {
		return err
	}

	return nil
}
