package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	wordfiles = flag.String("wordfiles", "adjectives.txt animals.txt", "Files with the link words")
	logfile   = flag.String("log", "linkto.log", "Log file path")

	hostname  = flag.String("host", "", "Linkto's hostname")
	redisPort = flag.String("rport", ":6379", "Redis' port")
)

const (
	serverPort = ":9091"

	passwordEnv = "LINKTO_PASSWORD"
)

func shorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")

	existing, err := longmap.Get(longurl)
	if err == nil {
		w.Write([]byte(fmt.Sprintf("Short link from %s to %s/%s exists\n", longurl, *hostname, existing)))
		return
	}
	shorturl := shortener.GetShortURL()
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

	c, err := setupRedis(*redisPort)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	files := strings.Split(*wordfiles, " ")
	err = setupShortener(shortmap, files)
	if err != nil {
		log.Fatalln(err)
	}

	logwriter, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logwriter)

	http.HandleFunc("/shorten", logreq(mustParams(shorten, "longurl")))
	http.HandleFunc("/customshorten", logreq(mustParams(simpleAuth(customshorten), "longurl", "customurl")))
	http.HandleFunc("/", logreq(redirect))
	log.Fatalln(http.ListenAndServe(serverPort, nil))
}
