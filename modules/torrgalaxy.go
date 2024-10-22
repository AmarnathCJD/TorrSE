package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TGX struct{}

func NewTGX() *TGX {
	return &TGX{}
}

// https://torrentgalaxy.mx/torrents.php?search=%s#results
// Actual TorrentGalaxy, But Slow

func (t *TGX) Search(query string, page int) ([]map[string]string, error) {
	startTime := time.Now()

	searchURL := fmt.Sprintf("https://torrentgalaxy.one/get-posts/keywords:%s", url.QueryEscape(query))

	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", getUA())
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

	docx, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := docx.Find("div.tgxtable")
	if results.Length() == 0 {
		return nil, nil
	}

	var parsedResults []map[string]string
	var sizes []string

	results.First().Find("div.tgxtablerow .txlight").Each(func(i int, doc *goquery.Selection) {
		title := doc.Find(".clickable-row a").AttrOr("title", "")
		link := doc.Find(".clickable-row a").AttrOr("href", "Link not found")
		size := doc.Find("span.badge").Text()

		if size != "" && strings.TrimSpace(size) != "1" {
			// if not last element in the array
			last := len(sizes) - 1
			if last < 0 {
				sizes = append(sizes, size)
			} else {
				if !reflect.DeepEqual(sizes[last], size) {
					sizes = append(sizes, size)
				}
			}
		}

		if title == "" || title == "Trusted Uploader" {
			return
		}
		parsedResults = append(parsedResults, map[string]string{
			"title":    title,
			"link":     "https://torrentgalaxy.one" + link,
			"provider": "TorrentGalaxy",
		})
	})

	for i, r := range parsedResults {
		if i >= len(sizes) {
			break
		}
		r["size"] = sizes[i]
	}

	logInfo(fmt.Sprintf("TGX search took %.2f seconds (%d)", time.Since(startTime).Seconds(), len(parsedResults)))
	return parsedResults, nil
}

func (t *TGX) GetMagnet(link string) (map[string]string, error) {
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
	doc.Find("a.btn-danger").Each(func(i int, result *goquery.Selection) {
		magnet, _ := result.Attr("href")
		magnets = append(magnets, map[string]string{
			"magnet": magnet,
		})
	})

	// torrentpagetable
	doc.Find(".torrentpagetable").Each(func(i int, result *goquery.Selection) {
		size := result.Find(".tprow").Eq(5).Find(".tpcell")
		sz := strings.Split(size.Text(), "Total Size:")
		if len(sz) == 1 {
			magnets[0]["size"] = "Unknown"
			return
		}
		magnets[0]["size"] = strings.TrimSpace(strings.Split(size.Text(), "Total Size:")[1])
	})

	return magnets[0], nil
}
