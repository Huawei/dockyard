package redis

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"sync"
	"time"

	"gopkg.in/redis.v3"
)

const luaRefresh = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`
const luaRelease = `if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`

type Lock struct {
	client *redis.Client
	key    string
	ttl    string
	opts   *LockOptions

	token string
	mutex sync.Mutex
}

// ObtainLock is a shortcut for NewLock().Lock()
func ObtainLock(client *redis.Client, key string, opts *LockOptions) (*Lock, error) {
	lock := NewLock(client, key, opts)
	if ok, err := lock.Lock(); err != nil || !ok {
		return nil, err
	}
	return lock, nil
}

// NewLock creates a new distributed lock on key
func NewLock(client *redis.Client, key string, opts *LockOptions) *Lock {
	opts = opts.normalize()
	ttl := strconv.FormatInt(int64(opts.LockTimeout/time.Millisecond), 10)
	return &Lock{client: client, key: key, ttl: ttl, opts: opts}
}

// IsLocked returns true if a lock is acquired
func (l *Lock) IsLocked() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.token != ""
}

// Lock applies the lock, don't forget to defer the Unlock() function to release the lock after usage
func (l *Lock) Lock() (bool, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.token != "" {
		return l.refresh()
	}
	return l.create()
}

// Unlock releases the lock
func (l *Lock) Unlock() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.release()
}

// Helpers
func (l *Lock) create() (bool, error) {
	l.reset()

	// Create a random token
	token, err := randomToken()
	if err != nil {
		return false, err
	}

	// Calculate the timestamp we are willing to wait for
	stop := time.Now().Add(l.opts.WaitTimeout)
	for {
		// Try to obtain a lock
		ok, err := l.obtain(token)
		if err != nil {
			return false, err
		} else if ok {
			l.token = token
			return true, nil
		}

		if time.Now().Add(l.opts.WaitRetry).After(stop) {
			break
		}
		time.Sleep(l.opts.WaitRetry)
	}
	return false, nil
}

func (l *Lock) refresh() (bool, error) {
	status, err := l.client.Eval(luaRefresh, []string{l.key}, []string{l.token, l.ttl}).Result()
	if err != nil {
		return false, err
	} else if status == int64(1) {
		return true, nil
	}
	return l.create()
}

func (l *Lock) obtain(token string) (bool, error) {
	cmd := redis.NewStringCmd("set", l.key, token, "nx", "px", l.ttl)
	l.client.Process(cmd)

	str, err := cmd.Result()
	if err == redis.Nil {
		err = nil
	}
	return str == "OK", err
}

func (l *Lock) release() error {
	defer l.reset()

	err := l.client.Eval(luaRelease, []string{l.key}, []string{l.token}).Err()
	if err == redis.Nil {
		err = nil
	}
	return err
}

func (l *Lock) reset() {
	l.token = ""
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
