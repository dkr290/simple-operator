package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var v string

func handler(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, "<h1 style='font-size:24px;'>Version: %s</h1>\n", v); err != nil {
		log.Fatal(err)
	}
}

func handlerv(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintf(w, "<h1 style='font-size:24px;'>Version: %s</h1>\n", v); err != nil {
		log.Fatal(err)
	}
}

func main() {
	v = os.Getenv("VERSION")
	if v == "" {
		v = "unknown"
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/"+v, handlerv)
	port := ":9090"
	fmt.Println("Server running on", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
