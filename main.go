package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Mock struct {
	statusCode  int
	contentType string
	body        string
	filepath    string
}

func (m *Mock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dumpRequest(r)
	var body io.Reader

	if m.filepath != "" {
		file, err := os.Open(m.filepath)
		if err != nil {
			log.Print(err)

			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Internal Server Error\n")

			return
		}
		body = file
	} else {
		body = strings.NewReader(m.body)
	}

	w.Header().Set("Content-Type", m.contentType)
	w.WriteHeader(m.statusCode)

	if _, err := io.Copy(w, body); err != nil {
		log.Print(err)
	}
}

func dumpRequest(r *http.Request) {
	var w = os.Stdout

	fmt.Fprintf(w, "[%s] %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr)
	fmt.Fprintf(w, "%s %s %s\n", r.Method, r.URL, r.Proto)
	for key, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", key, value)
		}
	}
	// TODO: pretty-print if it's JSON
	fmt.Fprintf(w, "\n")
	if _, err := io.Copy(os.Stdout, r.Body); err != nil {
		log.Print(err)
	}
	fmt.Fprintf(w, "\n")
}

var statusCode int
var contentType string
var body string
var filepath string
var addr string

func init() {
	flag.IntVar(&statusCode, "code", 200, "status code")
	flag.StringVar(&contentType, "content-type", "text/plain; charset=utf-8", "")
	flag.StringVar(&body, "body", "It's Works!\n", "response body")
	flag.StringVar(&filepath, "file", "", "file path to use as a response body")
	flag.StringVar(&addr, "addr", ":8080", "address to serve")
}

func main() {
	flag.Parse()

	m := Mock{
		statusCode:  statusCode,
		contentType: contentType,
		body:        body,
		filepath:    filepath,
	}

	log.Printf("serve at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, &m))
}
