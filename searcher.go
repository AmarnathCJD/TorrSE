package main

import (
	"encoding/json"
	"fmt"
	"main/modules"
	"net/http"
	"sync"
	"time"
)

type searchResult struct {
	sync.RWMutex
	Results []map[string]string
}

type finResult struct {
	Results    []map[string]string `json:"results"`
	Time       string              `json:"time"`
	TotalCount int                 `json:"total_count"`
}

func (s *searchResult) AddResults(results []map[string]string) {
	s.Lock()
	defer s.Unlock()
	s.Results = append(s.Results, results...)
}

func searchAsyncForTorrent(query string, w http.ResponseWriter, intent bool) {
	a, b, c, d, e, f := modules.NewTGX(), modules.NewYTS(), modules.NewTTIP(), modules.NewTor(), modules.New1337(), modules.NewTPB()

	var finalResult searchResult
	startTime := time.Now()
	wg := sync.WaitGroup{}

	searchAndAddResults := func(searchFunc func(string, int) ([]map[string]string, error), query string) {
		defer wg.Done()
		r, err := searchFunc(query, 1)
		if err != nil {
			modules.LogGlobalError(err)
			return
		}
		finalResult.AddResults(r)
	}

	wg.Add(6)
	go searchAndAddResults(a.Search, query)
	go searchAndAddResults(b.Search, query)
	go searchAndAddResults(c.Search, query)
	go searchAndAddResults(d.Search, query)
	go searchAndAddResults(e.Search, query)
	go searchAndAddResults(f.Search, query)

	wg.Wait()
	searchTime := time.Since(startTime).String()

	w.Header().Set("Content-Type", "application/json")
	if intent {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	enc := json.NewEncoder(w)
	if intent {
		enc.SetIndent("", "  ")
	}

	if err := enc.Encode(finResult{
		Results:    finalResult.Results,
		Time:       searchTime,
		TotalCount: len(finalResult.Results),
	}); err != nil {
		http.Error(w, `{"error": "failed to encode json"}`, http.StatusInternalServerError)
		return
	}

	modules.LogGlobalInfo(fmt.Sprintf("searched for %s in %s", query, searchTime))
}
