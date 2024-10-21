package main

import (
	"log"
	"main/modules"
	"net/http"
)

func main() {
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			w.Write([]byte("query not found"))
			return
		}

		searchAsyncForTorrent(query, w, r.URL.Query().Get("i") == "true")
	})

	modules.LogGlobalInfo("server started on :8080")
	log.Fatal(http.ListenAndServe(":80", nil))
}
