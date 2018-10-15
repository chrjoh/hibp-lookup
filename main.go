package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	requestPath = "/api/breachedaccount/"

	emailFile  = "email-list.txt"
	resultFile = "email-list-result.json"
	// Command line flags
	inputFunc  = flag.String("i", emailFile, "Input file with one email per line")
	outputFunc = flag.String("o", resultFile, "Output file with the result")
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nCommand line arguments:\n\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Parse the command line flags
	flag.Parse()
	handleQueries(*inputFunc)
}

func handleQueries(filePath string) {
	file, err := os.Open(filePath)
	ba := BreachedAccounts{}
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	h := newHIBP()
	for scanner.Scan() {
		email := scanner.Text()
		fmt.Printf("Checking email: %s\n", email)
		if result, err := h.GetBreaches(email); err == nil {
			if err != nil {
				log.Fatal(err)
			}
			if len(result.Breaches) > 0 {
				ba.Accounts = append(ba.Accounts, *result)
			}
		}
		time.Sleep(2 * time.Second)
	}
	hibpJson, _ := json.Marshal(ba)
	err = ioutil.WriteFile(*outputFunc, hibpJson, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

type HIBP struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func newHIBP() *HIBP {
	return &HIBP{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    &url.URL{Scheme: "https", Host: "haveibeenpwned.com"},
	}
}

type Breaches struct {
	Breaches []*string `json:"breaches"`
	Email    string    `json:"email"`
}

type BreachedAccounts struct {
	Accounts []Breaches `json:"accounts"`
}

func (h *HIBP) GetBreaches(email string) (*Breaches, error) {
	resp, err := h.get(fmt.Sprintf("%s%s", requestPath, url.PathEscape(email)))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	s := []*string{}
	json.NewDecoder(resp.Body).Decode(&s)

	return &Breaches{Breaches: s, Email: email}, nil
}

func (h *HIBP) get(path string) (*http.Response, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	return h.httpClient.Get(h.baseURL.ResolveReference(ref).String())
}
