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
)

var (
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

func createShortURL() string {
	var url string
	var err error
	for err == nil {
		url = fmt.Sprintf("%s%s", adjectives[rand.Intn(len(adjectives))], animals[rand.Intn(len(animals))])
		_, err = shortmap.Get(url)
		if err == nil {
			log.Printf("Created shortlink %s but was already present in the map\n", url)
		}
	}
	return url
}

func shorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")

	existing, err := longmap.Get(longurl)
	if err == nil {
		w.Write([]byte(fmt.Sprintf("Short link from %s to %s/%s exists\n", longurl, *hostname, existing)))
		return
	}
	shorturl := createShortURL()
	longmap.Set(longurl, shorturl)
	shortmap.Set(shorturl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, *hostname, shorturl)))
}

func customshorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")
	customurl := r.Form.Get("customurl")

	_, err := custommap.Get(customurl)
	if err == nil {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(fmt.Sprintf("Custom URL %s/%s is already taken\n", *hostname, customurl)))
		return
	}
	custommap.Set(customurl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, *hostname, customurl)))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:len(r.URL.Path)]
	longurl, err := shortmap.Get(path)
	if err == nil {
		http.Redirect(w, r, longurl, http.StatusMovedPermanently)
		return
	}
	longurl, err = custommap.Get(path)
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

	c, err := setupRedis(*redisPort)
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
