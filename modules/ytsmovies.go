package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type YTS struct{}

func NewYTS() *YTS {
	return &YTS{}
}

func (y *YTS) Search(query string, page int) ([]map[string]string, error) {
	startTime := time.Now()

	searchURL := fmt.Sprintf("https://en.ytsmx.mx/browse-movies/%s/all/all/0/0/latest", url.QueryEscape(query))

	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

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

	results := doc.Find("div.__main.mClearfix")
	if results.Length() == 0 {
		return nil, nil
	}

	var parsedResults []map[string]string

	results.Find("div.card").Each(func(i int, result *goquery.Selection) {
		title, _ := result.Find("a").Attr("title")
		rating := strings.ReplaceAll(result.Find("h4").Text(), "\r\n", "")
		rating = strings.ReplaceAll(rating, " ", "")
		rating = strings.SplitAfter(rating, "/10")[0]
		rating = strings.ReplaceAll(rating, "\n", "")

		link, _ := result.Find("a").Attr("href")
		link = "https://en.ytsmx.mx" + link

		parsedResults = append(parsedResults, map[string]string{
			"title":    title,
			"rating":   rating,
			"link":     link,
			"provider": "YTS.mx",
		})
	})

	logInfo(fmt.Sprintf("YTS search took %.2f seconds", time.Since(startTime).Seconds()))
	return parsedResults, nil
}

func (y *YTS) GetMagnet(link string) (map[string][]map[string]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

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

	var magnets []map[string]string

	doc.Find("a.download-torrent").Each(func(i int, s *goquery.Selection) {
		magnetLink, exists := s.Attr("href")
		if exists {
			title := strings.ReplaceAll(s.Text(), "\r\n", "")
			title = strings.ReplaceAll(title, " ", "")
			title = strings.ReplaceAll(title, "\n", "")
			title = strings.ReplaceAll(title, "º", "")

			magnets = append(magnets, map[string]string{
				"magnet": magnetLink,
				"title":  title,
			})
		}
	})

	return map[string][]map[string]string{"magnets": magnets}, nil
}
