package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
)

// shortener creates short strings based on the words it was initialised with
type shortener struct {
	wordsSlice [][]string
	existing   StringStore
}

func (s shortener) getShortURL() string {
	var url string
	var err error
	for err == nil {
		url = s.createRandomString()
		_, err = s.existing.Get(url)
		if err == nil {
			log.Printf("Created shortlink %s but was already present in the map\n", url)
		}
	}
	return url
}

func (s shortener) createRandomString() string {
	var bytes bytes.Buffer
	for _, words := range s.wordsSlice {
		bytes.WriteString(words[rand.Intn(len(words))])
	}
	return bytes.String()
}

func newShortener(existing StringStore, files []string) (shortener, error) {
	rand.Seed(time.Now().UnixNano())

	wordsSlice := make([][]string, len(files))
	for i, file := range files {
		words, err := readWords(file)
		if err != nil {
			return shortener{}, err
		}
		wordsSlice[i] = words
	}

	return shortener{wordsSlice, existing}, nil
}

func readWords(filename string) ([]string, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return []string{}, err
	}
	words := strings.Split(string(d), "\n")
	return words, nil
}
