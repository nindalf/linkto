package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
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
	return redis.String(c.Do("HGET", tablename, key))
}

func shorten(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form, r.PostForm)
	longurl := r.Form.Get("longurl")
	if len(longurl) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	existing, err := redisGet(longmap, longurl)
	fmt.Println(existing, err)
	if err == nil {
		w.Write([]byte(existing))
		return
	}
	shorturl := getShortURL()
	redisSet(longmap, longurl, shorturl)
	redisSet(shortmap, shorturl, longurl)
	w.Write([]byte(shorturl))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	longurl, err := redisGet(shortmap, r.URL.Path[1:len(r.URL.Path)])
	fmt.Println(r.URL.Path, err)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(longurl)
}

func main() {
	animals, _ = readWords("animals4.txt")
	adjectives, _ = readWords("adjectives3.txt")
	rand.Seed(time.Now().UnixNano())

	var err error
	c, err = redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	http.HandleFunc("/shorten", shorten)
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":9091", nil)

}
