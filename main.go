package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	c redis.Conn

	adjectives []string
	animals    []string

	adjectivesfile = flag.String("adjectives", "adjectives.txt", "Adjectives file path")
	animalsfile    = flag.String("animals", "animals.txt", "Animals file path")
	logfile        = flag.String("log", "linkto.log", "Log file path")

	hostname   = flag.String("host", "", "Linkto's hostname")
	serverPort = flag.String("lport", ":9091", "Linkto's port")
	redisPort  = flag.String("rport", ":6379", "Redis' port")
)

const (
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
		w.Write([]byte(fmt.Sprintf("Short link from %s to %s/%s exists\n", longurl, *hostname, existing)))
		return
	}
	shorturl := createShortURL()
	redisSet(longmap, longurl, shorturl)
	redisSet(shortmap, shorturl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, *hostname, shorturl)))
}

func customshorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")
	customurl := r.Form.Get("customurl")

	_, err := redisGet(custommap, customurl)
	if err == nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(fmt.Sprintf("Custom URL %s/%s is already taken\n", *hostname, customurl)))
		return
	}
	redisSet(custommap, customurl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, *hostname, customurl)))
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
	flag.Parse()
	var err error

	animals, err = readWords(*animalsfile)
	if err != nil {
		log.Fatalln(err)
	}
	adjectives, err = readWords(*adjectivesfile)
	if err != nil {
		log.Fatalln(err)
	}

	c, err = redis.Dial("tcp", *redisPort)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	rand.Seed(time.Now().UnixNano())

	logwriter, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logwriter)

	http.HandleFunc("/shorten", logreq(mustParams(shorten, "longurl")))
	http.HandleFunc("/customshorten", logreq(mustParams(simpleAuth(customshorten), "longurl", "customurl")))
	http.HandleFunc("/", logreq(redirect))
	log.Fatalln(http.ListenAndServe(*serverPort, nil))
}
