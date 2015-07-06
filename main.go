package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	rateLimitSecs     = 10
	rateLimitRequests = 10

	redisPortEnv = "REDIS_PORT"
	hostnameEnv  = "LINKTO_HOSTNAME"
	wordfilesEnv = "LINKTO_WORDFILES"
	passwordEnv  = "LINKTO_PASSWORD"

	linktoPort = ":9091"
)

type response struct {
	Longurl  string `json:"longurl"`
	Shorturl string `json:"shorturl"`
}

type handler struct {
	long      StringStore
	short     StringStore
	custom    StringStore
	hostname  string
	shortener shortener
}

func newHandler(hostname string, s shortener) handler {
	long := newStringStore(longmapName)
	short := newStringStore(shortmapName)
	custom := newStringStore(custommapName)
	return handler{long, short, custom, hostname, s}
}

func (h handler) shorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")

	existing, err := h.long.Get(longurl)
	if err == nil {
		w.Header().Set("Content-Type", "text/json")
		w.Write(respJSON(longurl, h.hostname, existing))
		return
	}
	shorturl := h.shortener.getShortURL()
	h.long.Set(longurl, shorturl)
	h.short.Set(shorturl, longurl)

	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respJSON(longurl, h.hostname, shorturl))
}

func (h handler) customShorten(w http.ResponseWriter, r *http.Request) {
	longurl := r.Form.Get("longurl")
	customurl := r.Form.Get("customurl")

	_, errcu := h.custom.Get(customurl)
	_, errsh := h.short.Get(customurl)
	if errcu == nil || errsh == nil {
		fmt.Fprintf(w, "Custom URL %s/%s is already taken\n", h.hostname, customurl)
		return
	}
	h.custom.Set(customurl, longurl)

	w.Header().Set("Content-Type", "text/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respJSON(longurl, h.hostname, customurl))
}

func respJSON(longurl, hostname, shorturl string) []byte {
	s := fmt.Sprintf("%s/%s", hostname, shorturl)
	r := response{Longurl: longurl, Shorturl: s}
	body, _ := json.Marshal(r)
	return body
}

func (h handler) redirect(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:len(r.URL.Path)]
	if path == "" {
		w.Write([]byte(mainpage))
		return
	}
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
	http.Error(w, fmt.Sprintf("No link found for %s", path), http.StatusNotFound)
}

func main() {
	redisPort := os.Getenv(redisPortEnv)
	c, err := setupRedis(redisPort[strings.LastIndex(redisPort, "/")+1 : len(redisPort)])
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	files := strings.Split(os.Getenv(wordfilesEnv), " ")
	shortener, err := newShortener(newStringStore(shortmapName), files)
	if err != nil {
		log.Fatalln(err)
	}

	hostname := os.Getenv(hostnameEnv)
	h := newHandler(hostname, shortener)

	shorten := MustParams(RateLimit(h.shorten, rateLimitSecs, rateLimitRequests, newExpireStore()), "longurl")
	http.HandleFunc("/shorten", LogResp(CORS(shorten)))

	customShorten := MustParams(SimpleAuth(h.customShorten), "longurl", "customurl")
	http.HandleFunc("/customshorten", LogResp(CORS(customShorten)))

	http.HandleFunc("/", LogResp(h.redirect))

	log.Fatalln(http.ListenAndServe(linktoPort, nil))
}
