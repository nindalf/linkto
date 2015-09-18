package main

import (
	"errors"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	longmapName   = "longToShort"
	shortmapName  = "shortToLong"
	custommapName = "customToLong"
)

var (
	pool redis.Pool
)

// StringStore is a map from string to string
type StringStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

func newStringStore(name string) StringStore {
	return &redisMap{pool: pool, mapName: name}
}

type redisMap struct {
	pool    redis.Pool
	mapName string
}

func (r *redisMap) Get(key string) (string, error) {
	conn := r.pool.Get()
	defer conn.Close()
	return redis.String(conn.Do("HGET", r.mapName, key))
}

func (r *redisMap) Set(key, value string) error {
	conn := r.pool.Get()
	defer conn.Close()
	_, err := conn.Do("HSET", r.mapName, key, value)
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
	return &redisExpire{pool: pool}
}

type redisExpire struct {
	pool redis.Pool
}

func (r *redisExpire) Get(key string) (int, error) {
	conn := r.pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("GET", key))
}

func (r *redisExpire) Incr(key string, expireSecs int) error {
	conn := r.pool.Get()
	defer conn.Close()
	conn.Send("INCR", key)
	conn.Send("EXPIRE", key, expireSecs)
	conn.Flush()
	_, err := conn.Receive()
	return err
}

func (r *redisExpire) TTL(key string) (int, error) {
	conn := r.pool.Get()
	defer conn.Close()
	ttl, err := redis.Int(conn.Do("TTL", key))
	if err != nil {
		return ttl, err
	}
	if ttl < -1 {
		return ttl, errors.New("Key does not exist")
	}
	return ttl, nil
}

func setupRedis(port string) error {
	pool = redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", port)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	conn := pool.Get()
	_, err := conn.Do("PING")
	return err
}
