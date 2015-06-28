package main

import (
	"errors"

	"github.com/garyburd/redigo/redis"
)

const (
	longmapName   = "longToShort"
	shortmapName  = "shortToLong"
	custommapName = "customToLong"
)

var (
	c redis.Conn
)

// StringStore is a map from string to string
type StringStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

func newStringStore(name string) StringStore {
	return redisMap{c: c, mapName: name}
}

type redisMap struct {
	c       redis.Conn
	mapName string
}

func (r redisMap) Get(key string) (string, error) {
	return redis.String(r.c.Do("HGET", r.mapName, key))
}

func (r redisMap) Set(key, value string) error {
	_, err := r.c.Do("HSET", r.mapName, key, value)
	return err
}

// ExpireStore is a map from string to int.
// Keys can be incremented and are expired automatically.
type ExpireStore interface {
	Get(key string) (int, error)
	Incr(key string, expiry int) error
	TTL(key string) (int, error)
}

func newExpireStore() ExpireStore {
	return redisExpire{c: c}
}

type redisExpire struct {
	c redis.Conn
}

func (r redisExpire) Get(key string) (int, error) {
	return redis.Int(r.c.Do("GET", key))
}

func (r redisExpire) Incr(key string, expireSecs int) error {
	r.c.Send("INCR", key)
	r.c.Send("EXPIRE", key, expireSecs)
	r.c.Flush()
	_, err := r.c.Receive()
	return err
}

func (r redisExpire) TTL(key string) (int, error) {
	ttl, err := redis.Int(r.c.Do("TTL", key))
	if err != nil {
		return ttl, err
	}
	if ttl < -1 {
		return ttl, errors.New("Key does not exist")
	}
	return ttl, nil
}

func setupRedis(port string) (redis.Conn, error) {
	var err error
	c, err = redis.Dial("tcp", port)
	return c, err
}
