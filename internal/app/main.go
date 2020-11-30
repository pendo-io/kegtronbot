package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/appengine"
)

// indexHandler responds to requests with our greeting.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, World!")
}

func main() {
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	isDev := os.Getenv("ISDEV")
	isDev = strings.ToLower(isDev)
	if isDev == "y" || isDev == "yes" {
		http.ListenAndServe(":8080", nil)
	} else {
		appengine.Main()
	}

}