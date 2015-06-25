package main

import (
	"github.com/garyburd/redigo/redis"
)

const (
	longmapName   = "longToShort"
	shortmapName  = "shortToLong"
	custommapName = "customToLong"
)

var (
	longmap   KVStore
	shortmap  KVStore
	custommap KVStore
)

// KVStore abstracts the underlying storage medium
type KVStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type redisStore struct {
	c         redis.Conn
	tablename string
}

func (r redisStore) Get(key string) (string, error) {
	return redis.String(r.c.Do("HGET", r.tablename, key))
}

func (r redisStore) Set(key, value string) error {
	_, err := r.c.Do("HSET", r.tablename, key, value)
	return err
}

func setupRedis(tcpPort string) (redis.Conn, error) {
	c, err := redis.Dial("tcp", tcpPort)
	if err != nil {
		return nil, err
	}
	longmap = redisStore{c: c, tablename: longmapName}
	shortmap = redisStore{c: c, tablename: shortmapName}
	custommap = redisStore{c: c, tablename: custommapName}
	return c, err
}
