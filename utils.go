package main

import (
	"encoding/xml"
	"errors"
	"net/http"
)

type xmlSitemap struct {
	URLS []xmlURL `xml:"url"`
}

type xmlURL struct {
	Loc string `xml:"loc"`
}

// ParseSitemap - requests a sitemap and returns all the URLs from it.
func ParseSitemap(crawling *Crawling) ([]string, error) {
	c := http.Client{}
	result := []string{}

	urlSet := &xmlSitemap{}
	req, err := http.NewRequest("GET", crawling.URL, nil)
	if err != nil {
		return result, err
	}

	for _, header := range crawling.Headers {
		req.Header.Add(header.Name, header.Value)
	}

	resp, err := c.Do(req)
	if err != nil {
		return result, err
	}

	xml.NewDecoder(resp.Body).Decode(urlSet)

	for _, url := range urlSet.URLS {
		result = append(result, url.Loc)
	}
	if len(result) == 0 {
		return result, errors.New("No URLs can be parsed from the provided resource")
	}
	return result, nil
}
