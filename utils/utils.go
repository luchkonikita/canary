package utils

import (
	"encoding/xml"
	"net/http"
)

// URL is a structure of <url> in <sitemap>
type url struct {
	Loc string `xml:"loc"`
}

// ParseSitemap - requests a sitemap and returns all the URLs from it.
func ParseSitemap(sitemapURL string) ([]string, error) {
	c := http.Client{}
	result := []string{}

	urlSet := &struct {
		URLS []url `xml:"url"`
	}{}
	req, err := http.NewRequest("GET", sitemapURL, nil)
	if err != nil {
		return result, err
	}

	// if sitemap.HasAuth() {
	// 	req.SetBasicAuth(sitemap.Username, sitemap.Password)
	// }

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
