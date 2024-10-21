package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type TPB struct{}

func NewTPB() *TPB {
	return &TPB{}
}

func (t *TPB) Search(query string) ([]map[string]string, error) {
	startTime := time.Now()

	encodedQuery := url.QueryEscape(query)
	searchURL := "https://tpirbay.site/search/" + encodedQuery + "/0/99/0"

	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
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

	results := doc.Find("table#searchResult tr").NextAll()
	var parsedResults []map[string]string

	results.Each(func(i int, result *goquery.Selection) {
		metaText := result.Find("td").Eq(1).Text()
		meta := strings.Split(metaText, "\n")
		uploaded := ""
		if len(meta) > 4 {
			uploaded = strings.ReplaceAll(meta[3], "\u00a0", " ")
		}

		// re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})`)
		// uploadTimeMatch := re.FindStringSubmatch(uploaded)
		// uploadTime := ""
		// if len(uploadTimeMatch) > 0 {
		// 	uploadTime = uploadTimeMatch[0]
		// }

		uploadedBy := ""
		if len(strings.Split(uploaded, ",")) > 2 {
			uploadedBy = strings.TrimSpace(strings.Split(uploaded, ",")[2])
		}

		name := result.Find("a.detLink").Text()
		magnet, _ := result.Find("a[title='Download this torrent using magnet']").Attr("href")
		size := ""
		if len(strings.Split(uploaded, ",")) > 1 {
			size = strings.TrimSpace(strings.Replace(strings.Split(uploaded, ",")[1], "Size ", "", 1))
		}

		seeders := result.Find("td").Eq(2).Text()
		leechers := result.Find("td").Eq(3).Text()

		parsedResults = append(parsedResults, map[string]string{
			"name":        name,
			"magnet":      magnet,
			"size":        size,
			"seeders":     seeders,
			"leechers":    leechers,
			"uploaded":    strings.TrimSpace(strings.Replace(uploaded, "Uploaded ", "", 1)),
			"uploaded_by": strings.TrimSpace(strings.Replace(uploadedBy, "by ", "", 1)),
			"provider":    "ThePirateBay",
		})
	})

	logInfo("TPB search took " + fmt.Sprintf("%.2f", time.Since(startTime).Seconds()) + " seconds")
	return parsedResults, nil
}
