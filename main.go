package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

var (
	animals    []string
	adjectives []string
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

func main() {
	animals, _ = readWords("animals4.txt")
	adjectives, _ = readWords("adjectives3.txt")
	rand.Seed(time.Now().UnixNano())
	for {
		<-time.After(time.Second)
		fmt.Println(getShortURL())
	}
}
