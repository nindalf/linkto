package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	c redis.Conn

	animals    []string
	adjectives []string

	longmap   = "longToShort"
	shortmap  = "shortToLong"
	custommap = "customToLong"
)

func readWords(filename string) ([]string, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}
	words := strings.Split(string(d), "\n")
	return words, nil
}

func getShortURL() string {
	return fmt.Sprintf("%s%s", adjectives[rand.Intn(len(adjectives))], animals[rand.Intn(len(animals))])
}

func redisSet(tablename, key, value string) error {
	_, err := c.Do("HSET", tablename, key, value)
	return err
}

func redisGet(tablename, key string) (string, error) {
	redis.String(c.Do("HGET", tablename, key))
}

func main() {
	animals, _ = readWords("animals4.txt")
	adjectives, _ = readWords("adjectives3.txt")
	rand.Seed(time.Now().UnixNano())

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

}
