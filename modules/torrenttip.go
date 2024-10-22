package modules

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TTIP struct{}

func NewTTIP() *TTIP {
	return &TTIP{}
}

func (t *TTIP) Search(query string, page int) ([]map[string]string, error) {
	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://torrenttip146.com/search", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("page", fmt.Sprintf("%d", page))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("User-Agent", getUA())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := doc.Find("ul.page-list")
	if results.Length() == 0 {
		return nil, nil
	}

	var parsedResults []map[string]string

	results.Find("li").Each(func(i int, s *goquery.Selection) {
		name := s.Find("a").Text()
		link, _ := s.Find("a").Attr("href")
		uploaded := s.Find("div.text-right").Text()

		parsedResults = append(parsedResults, map[string]string{
			"name":     name,
			"link":     link,
			"uploaded": uploaded,
			"provider": "TorrentTip",
		})
	})

	logInfo(fmt.Sprintf("TTIP search took %.2f seconds (%d)", time.Since(startTime).Seconds(), len(parsedResults)))
	return parsedResults, nil
}

func (t *TTIP) GetMagnet(link string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", getUA())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	downloadBtn := doc.Find("a.ml-2.border.bg-main-color.text-white.py-2.pr-2.text-16px")
	if downloadBtn.Length() == 0 {
		return "", fmt.Errorf("magnet link not found")
	}

	magnetLink, _ := downloadBtn.Attr("href")
	magnetLink = strings.TrimSpace(magnetLink)
	return magnetLink, nil
}
