package utils

import (
	"encoding/xml"
	"net/http"

	"github.com/luchkonikita/canary/store"
)

// URL is a structure of <url> in <sitemap>
type URL struct {
	Loc string `xml:"loc"`
}

// ParseSitemap - requests a sitemap and returns all the URLs from it.
func ParseSitemap(sitemap *store.Sitemap) ([]string, error) {
	c := http.Client{}
	result := []string{}

	urlSet := &struct {
		URLS []URL `xml:"url"`
	}{}
	req, err := http.NewRequest("GET", sitemap.URL, nil)
	if err != nil {
		return result, err
	}

	if sitemap.HasAuth() {
		req.SetBasicAuth(sitemap.Username, sitemap.Password)
	}

	resp, err := c.Do(req)
	if err != nil {
		return result, err
	}

	xml.NewDecoder(resp.Body).Decode(urlSet)

	for _, url := range urlSet.URLS {
		result = append(result, url.Loc)
	}

	return result, nil
}
