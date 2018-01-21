package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

const queryLimit = 20000

// Filters
// For building advanced queries and using parameters

// CrawlingsFilter a struct used for filtering crawlings based on JSON encoded parameters
type CrawlingsFilter struct {
	Filter *FilterQuery
}

func NewCrawlingsFilter(r *http.Request) *CrawlingsFilter {
	return &CrawlingsFilter{Filter: NewFilterQuery(r)}
}

// Query - applies CrawlingsFilter and returns results
func (f *CrawlingsFilter) Query(db *storm.DB) []Crawling {
	crawlings := []Crawling{}
	filters := []q.Matcher{}

	if url, err := f.Filter.StringValue("url"); err == nil {
		filters = append(filters, q.Eq("URL", url))
	}
	if processed, err := f.Filter.BoolValue("processed"); err == nil {
		filters = append(filters, q.Eq("Processed", processed))
	}

	query := db.Select(q.And(filters...))
	query = query.Limit(f.Filter.LimitValue())
	query = query.Skip(f.Filter.OffsetValue())

	query.Find(&crawlings)
	return crawlings
}

// PageResultsFilter a struct used for filtering page results based on JSON encoded parameters
type PageResultsFilter struct {
	Filter *FilterQuery
}

func NewPageResultsFilter(r *http.Request) *PageResultsFilter {
	return &PageResultsFilter{Filter: NewFilterQuery(r)}
}

// Query - applies PageResultsFilter and returns results
func (f *PageResultsFilter) Query(db *storm.DB) []PageResult {
	pageResults := []PageResult{}
	filters := []q.Matcher{}

	if crawlingID, err := f.Filter.IntValue("crawling_id"); err == nil {
		filters = append(filters, q.Eq("CrawlingID", crawlingID))
	}
	if status, err := f.Filter.IntValue("status"); err == nil {
		filters = append(filters, q.Eq("Status", status))
	}
	if url, err := f.Filter.StringValue("url"); err == nil {
		filters = append(filters, q.Re("URL", url))
	}

	query := db.Select(q.And(filters...))
	query = query.Limit(f.Filter.LimitValue())
	query = query.Skip(f.Filter.OffsetValue())

	query.Find(&pageResults)
	return pageResults
}

type FilterQuery struct {
	Request *http.Request
}

func NewFilterQuery(r *http.Request) *FilterQuery {
	r.ParseForm()
	return &FilterQuery{Request: r}
}

func (fq *FilterQuery) StringValue(name string) (string, error) {
	val := fq.Request.FormValue(name)
	if val == "" {
		return val, errors.New("is empty")
	}
	return val, nil
}

func (fq *FilterQuery) IntValue(name string) (int, error) {
	val := fq.Request.FormValue(name)
	return strconv.Atoi(val)
}

func (fq *FilterQuery) BoolValue(name string) (bool, error) {
	val := fq.Request.FormValue(name)

	if val == "true" {
		return true, nil
	} else if val == "false" {
		return false, nil
	}
	return false, errors.New("is empty")
}

func (fq *FilterQuery) OffsetValue() int {
	offset, _ := fq.IntValue("offset")
	return offset
}

func (fq *FilterQuery) LimitValue() int {
	if limit, err := fq.IntValue("limit"); err == nil {
		if limit > 0 && limit < queryLimit {
			return limit
		}
		return queryLimit
	}
	return queryLimit
}
