package modules

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	siteName = "TorLock"
	baseURL  = "https://www.torlock.com"
)

type Tor struct{}

func NewTor() *Tor {
	return &Tor{}
}

func (t *Tor) Search(query string, page int) ([]map[string]string, error) {
	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("qq", "1")
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

	var parsedResults []map[string]string

	doc.Find("table.table-striped.table-bordered.table-hover.table-condensed").Each(func(i int, table *goquery.Selection) {
		if i == 1 {
			table.Find("tr").Each(func(index int, row *goquery.Selection) {
				if index == 0 {
					return
				}
				title := row.Find("a").Text()
				size := row.Find("td.ts").Text()
				seeds := row.Find("td.tul").Text()
				leeches := row.Find("td.tdl").Text()
				link, exists := row.Find("a").Attr("href")
				if !exists {
					return
				}
				uploaded := row.Find("td.td").Text()

				parsedResults = append(parsedResults, map[string]string{
					"title":    title,
					"size":     size,
					"seeds":    seeds,
					"leeches":  leeches,
					"link":     baseURL + link,
					"uploaded": uploaded,
					"provider": siteName,
				})
			})
		}
	})

	logInfo(fmt.Sprintf("TorLock search took %.2f seconds", time.Since(startTime).Seconds()))
	return parsedResults, nil
}

func (t *Tor) GetMagnet(link string) (string, error) {
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)

	regex := regexp.MustCompile(`magnet:\?xt=urn:btih:[0-9a-fA-F]{40}&dn=[^&]+(?:&tr=[^&]+)*`)
	matches := regex.FindStringSubmatch(bodyString)

	if len(matches) == 0 {
		return "", nil
	}

	magnet := strings.Split(matches[0], "\"")[0]
	return magnet, nil
}
