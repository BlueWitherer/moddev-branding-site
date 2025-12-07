package api

import (
	"fmt"
	"net/http"

	"service/log"
)

func init() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Mod Developer Branding API service pinged")
		header := w.Header()

		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Methods", "GET")
		header.Set("Access-Control-Allow-Headers", "Content-Type")
		header.Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong!")
	})
}
