package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	c redis.Conn

	redisPort  = ":6379"
	serverPort = ":9091"

	animals    []string
	adjectives []string

	longmap   = "longToShort"
	shortmap  = "shortToLong"
	custommap = "customToLong"

	passwordEnv = "LINKTOPASSWORD"
)

func readWords(filename string) ([]string, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}
	words := strings.Split(string(d), "\n")
	return words, nil
}

func redisSet(tablename, key, value string) error {
	_, err := c.Do("HSET", tablename, key, value)
	return err
}

func redisGet(tablename, key string) (string, error) {
	return redis.String(c.Do("HGET", tablename, key))
}

func createShortURL() string {
	var url string
	var err error
	for err == nil {
		url = fmt.Sprintf("%s%s", adjectives[rand.Intn(len(adjectives))], animals[rand.Intn(len(animals))])
		_, err = redisGet(shortmap, url)
		if err == nil {
			log.Printf("Created shortlink %s but was already present in the map\n", url)
		}
	}
	return url
}

func shorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")

	existing, err := redisGet(longmap, longurl)
	if err == nil {
		w.Write([]byte(fmt.Sprintf("Short link from %s to %s exists\n", longurl, existing)))
		return
	}
	shorturl := createShortURL()
	redisSet(longmap, longurl, shorturl)
	redisSet(shortmap, shorturl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s\n", longurl, shorturl)))
}

func customshorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")
	customurl := r.Form.Get("customurl")

	_, err := redisGet(custommap, customurl)
	if err == nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(fmt.Sprintf("Custom URL %s is already taken\n", customurl)))
		return
	}
	redisSet(custommap, customurl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s\n", longurl, customurl)))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:len(r.URL.Path)]
	longurl, err := redisGet(shortmap, path)
	if err == nil {
		http.Redirect(w, r, longurl, http.StatusMovedPermanently)
		return
	}
	longurl, err = redisGet(custommap, path)
	if err == nil {
		http.Redirect(w, r, longurl, http.StatusMovedPermanently)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func main() {
	var err error
	animals, err = readWords("animals.txt")
	if err != nil {
		log.Fatalln(err)
	}
	adjectives, err = readWords("adjectives.txt")
	if err != nil {
		log.Fatalln(err)
	}

	c, err = redis.Dial("tcp", redisPort)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/shorten", logreq(mustParams(shorten, "longurl")))
	http.HandleFunc("/customshorten", logreq(mustParams(simpleAuth(customshorten), "longurl", "customurl")))
	http.HandleFunc("/", logreq(redirect))
	http.ListenAndServe(serverPort, nil)
}
