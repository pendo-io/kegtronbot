package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pendo-io/pankbot/internal/app/pankbot"
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
	http.HandleFunc("/slack", pankbot.HandleSlack)
	http.HandleFunc("/slackInteractive", pankbot.HandleInteractivePankResponse)
	http.HandleFunc("/getPanks", pankbot.GetPanks)
	http.HandleFunc("/postPanks", pankbot.PostPanks)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	appengine.Main()
}

// auth store
// panks
// channels
