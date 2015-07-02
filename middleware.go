package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// LogResp logs all the responses to clients
// It does this by replacing the default ResponseWriter with a custom one
func LogResp(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddress(r)
		log.Printf("%s - Received req for %s\n", ip, r.URL.Path)
		logger := responseLogger{w: w, ip: ip}
		fn(logger, r)
	}
}

// MustParams checks for the existence of all specified params in the request
// If they don't exist, the request is declined as BadRequest (400)
func MustParams(fn http.HandlerFunc, params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		for _, param := range params {
			if len(r.Form.Get(param)) == 0 {
				errortext := fmt.Sprintf("Expected parameter %s not found", param)
				http.Error(w, errortext, http.StatusBadRequest)
				return
			}
		}
		fn(w, r)
	}
}

// SimpleAuth checks if a password has been set in the ENV and
// compares it to the request's password
// If they don't match, the request is declined as Unauthorized (401)
func SimpleAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linktopassword := os.Getenv(passwordEnv)
		if len(linktopassword) > 0 {
			password := r.Form.Get("password")
			if linktopassword != password {
				errortext := "Password param incorrect"
				http.Error(w, errortext, http.StatusUnauthorized)
				return
			}
		}
		fn(w, r)
	}
}

// RateLimit checks if the client making the request has been making
// more than numreq requests in the specified duration
// If it has, the request is declined as TooManyRequests (429)
func RateLimit(fn http.HandlerFunc, duration, numreq int, store ExpireStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddress(r)
		requests, err := store.Get(ip)
		if err != nil || requests < numreq {
			store.Incr(ip, duration)
			fn(w, r)
			return
		}
		ttl, _ := store.TTL(ip)
		errortext := fmt.Sprintf("You've sent too many requests in a short span of time. Try again after %d seconds", ttl)
		http.Error(w, errortext, 429)
	}
}

// CORS enables Cross-origin resource sharing for this request.
func CORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		fn(w, r)
	}
}

func getIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For") // assuming its behind nginx
	if len(ip) > 0 {
		return ip
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

type responseLogger struct {
	w  http.ResponseWriter
	ip string
}

func (r responseLogger) Header() http.Header {
	return r.w.Header()
}

// Write logs the first line of the response and writes the full response to the ResponseWriter
func (r responseLogger) Write(b []byte) (int, error) {
	logtext := string(b)
	endl := strings.IndexRune(logtext, '\n')
	if endl > 0 {
		if len(logtext) > endl+2 {
			logtext = logtext[0:endl] + " ... (truncated)\n"
		}
	} else {
		logtext = logtext + "\n"
	}
	log.Printf(fmt.Sprintf("%s - %s", r.ip, logtext))
	return r.w.Write(b)
}

func (r responseLogger) WriteHeader(code int) {
	log.Printf("%s - Returned code %d\n", r.ip, code)
	r.w.WriteHeader(code)
}
