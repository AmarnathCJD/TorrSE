package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type x1337 struct{}

func New1337() *x1337 {
	return &x1337{}
}

func (t *x1337) Search(query string, page int) ([]map[string]string, error) {
	fmt.Println("1337x search")
	startTime := time.Now()

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.1337xx.to/search/"+url.QueryEscape(query)+"/1/", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", getUA())

	resp, err := client.Do(req)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	fmt.Println(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsedResults []map[string]string

	doc.Find(".box-info-detail").First().Find("tr").Each(func(i int, s *goquery.Selection) {
		title := s.Find("td:nth-child(1)").Find("a").Last().Text()
		link, _ := s.Find("td:nth-child(1)").Find("a").Last().Attr("href")
		seeders := s.Find("td:nth-child(2)").Text()
		leechers := s.Find("td:nth-child(3)").Text()
		uploaded := s.Find("td:nth-child(4)").Text()
		size := s.Find("td:nth-child(5)").Text()

		if title == "" {
			return
		}

		parsedResults = append(parsedResults, map[string]string{
			"title":    title,
			"link":     "https://www.1337xx.to" + link,
			"seeders":  seeders,
			"leechers": leechers,
			"uploaded": uploaded,
			"size":     size,
			"provider": "1337x",
		})
	})

	logInfo(fmt.Sprintf("1337X search took %.2f seconds (%d)", time.Since(startTime).Seconds(), len(parsedResults)))
	return parsedResults, nil
}

func (t *x1337) GetMagnet(link string) (string, error) {
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

	magnet := doc.Find("a.torrentdown1").First().AttrOr("href", "")
	return magnet, nil
}
