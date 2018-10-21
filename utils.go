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

func requestSitemap(url string, headers []requestHeader) ([]string, error) {
	result := []string{}
	urlSet := &xmlSitemap{}

	res, err := requestWithRetries(url, headers, 3)
	if err != nil {
		return result, err
	}

	xml.NewDecoder(res.Body).Decode(urlSet)

	for _, url := range urlSet.URLS {
		result = append(result, url.Loc)
	}
	if len(result) == 0 {
		return result, errors.New("No URLs can be parsed from the provided resource")
	}
	return result, nil
}

func requestPage(url string, headers []requestHeader) (int, error) {
	res, err := requestWithRetries(url, headers, 3)
	if err != nil {
		return 0, err
	}
	return res.StatusCode, nil
}

func requestWithRetries(url string, headers []requestHeader, attempts int) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &http.Response{}, errors.New("Cannot perform the request")
	}

	for _, header := range headers {
		req.Header.Add(header.Name, header.Value)
	}

	res, err := client.Do(req)

	if err != nil {
		return &http.Response{}, errors.New("Cannot perform the request")
	}

	if res.StatusCode != 200 && attempts > 0 {
		return requestWithRetries(url, headers, attempts-1)
	}
	return res, err
}
