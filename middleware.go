package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func mustParams(fn http.HandlerFunc, params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		for _, param := range params {
			if len(r.Form.Get(param)) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("Expected parameter %s not found\n", param)))
				return
			}
		}
		fn(w, r)
	}
}

func simpleAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		linktopassword := os.Getenv(passwordEnv)
		if len(linktopassword) > 0 {
			password := r.Form.Get("password")
			if linktopassword != password {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Password param incorrect\n"))
				return
			}
		}
		fn(w, r)
	}
}

func logreq(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received req for " + r.URL.Path)
		logger := responseLogger{w: w}
		fn(logger, r)
	}
}

type responseLogger struct {
	w http.ResponseWriter
}

func (r responseLogger) Header() http.Header {
	return r.w.Header()
}

func (r responseLogger) Write(b []byte) (int, error) {
	log.Print(string(b))
	return r.w.Write(b)
}

func (r responseLogger) WriteHeader(code int) {
	log.Printf("Returned code %d\n", code)
	r.w.WriteHeader(code)
}
