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

	rateLimitSecs     = 60
	rateLimitRequests = 10

	passwordEnv = "LINKTO_PASSWORD"
)

type handler struct {
	long      StringStore
	short     StringStore
	custom    StringStore
	hostname  string
	shortener Shortener
}

func newHandler(hostname string, s Shortener) handler {
	long := newStringStore(longmapName)
	short := newStringStore(shortmapName)
	custom := newStringStore(custommapName)
	return handler{long, short, custom, hostname, s}
}

func (h handler) shorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")

	existing, err := h.long.Get(longurl)
	if err == nil {
		errortext := fmt.Sprintf("Short link from %s to %s/%s exists\n", longurl, h.hostname, existing)
		http.Error(w, errortext, http.StatusTeapot)
		return
	}
	shorturl := h.shortener.GetShortURL()
	h.long.Set(longurl, shorturl)
	h.short.Set(shorturl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, h.hostname, shorturl)))
}

func (h handler) customShorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")
	customurl := r.Form.Get("customurl")

	_, err := h.custom.Get(customurl)
	if err == nil {
		errortext := fmt.Sprintf("Custom URL %s/%s is already taken\n", h.hostname, customurl)
		http.Error(w, errortext, http.StatusTeapot)
		return
	}
	h.custom.Set(customurl, longurl)
	w.Write([]byte(fmt.Sprintf("Shortened %s to %s/%s\n", longurl, h.hostname, customurl)))
}

func (h handler) redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:len(r.URL.Path)]
	longurl, err := h.short.Get(path)
	if err == nil {
		http.Redirect(w, r, longurl, http.StatusMovedPermanently)
		return
	}
	longurl, err = h.custom.Get(path)
	if err == nil {
		http.Redirect(w, r, longurl, http.StatusMovedPermanently)
		return
	}
	http.Error(w, fmt.Sprintf("No link found for %s", path), http.StatusBadRequest)
}

func main() {
	flag.Parse()

	c, err := setupRedis(*redisPort)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	files := strings.Split(*wordfiles, " ")
	shortener, err := newShortener(newStringStore(shortmapName), files)
	if err != nil {
		log.Fatalln(err)
	}

	logwriter, err := os.OpenFile(*logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logwriter)

	h := newHandler(*hostname, shortener)

	shorten := MustParams(RateLimit(h.shorten, rateLimitSecs, rateLimitRequests, newExpireStore()), "longurl")
	http.HandleFunc("/shorten", LogResp(shorten))

	customShorten := MustParams(SimpleAuth(h.customShorten), "longurl", "customurl")
	http.HandleFunc("/customshorten", LogResp(customShorten))

	http.HandleFunc("/", LogResp(h.redirect))

	log.Fatalln(http.ListenAndServe(serverPort, nil))
}
