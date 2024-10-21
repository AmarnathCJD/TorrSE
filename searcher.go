package main

import (
	"fmt"
	"main/modules"
	"net/http"
	"time"
)

type searchResult struct {
	Results []map[string]string `json:"results"`
	Time    string              `json:"time"`
}

func searchAsyncForTorrent(query string, w http.ResponseWriter, intent bool) {
	a, b, c, d, e := modules.NewTGX(), modules.NewYTS(), modules.NewTTIP(), modules.NewTor(), modules.NewYTS()
	results := make(chan []map[string]string, 5)
	errors := make(chan error, 5)

	startTime := time.Now()

	go func() {
		r, err := a.Search(query, 1)
		results <- r
		errors <- err
	}()

	go func() {
		r, err := b.Search(query, 1)
		results <- r
		errors <- err
	}()

	go func() {
		r, err := c.Search(query, 1)
		results <- r
		errors <- err
	}()

	go func() {
		r, err := d.Search(query, 1)
		results <- r
		errors <- err
	}()

	go func() {
		r, err := e.Search(query, 1)
		results <- r
		errors <- err
	}()

	var finalResults []map[string]string

	for i := 0; i < 5; i++ {
		select {
		case result := <-results:
			if result != nil {
				finalResults = append(finalResults, result...)
			}
		case err := <-errors:
			if err != nil {
				modules.LogGlobalError(err)
			}
		}
	}

	modules.LogGlobalInfo(fmt.Sprintf("searched for %s in %s", query, time.Since(startTime)))
	w.Header().Set("Content-Type", "application/json")
	// add search time to the response

	var searchTime = time.Since(startTime).String()
	searchTime = searchTime[:len(searchTime)-7]

	searchResult := searchResult{
		Results: finalResults,
		Time:    searchTime,
	}

	modules.WriteJSON(w, searchResult, intent)
}
