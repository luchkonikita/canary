package store

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gorilla/schema"
	"github.com/luchkonikita/canary/utils"
)

const queryLimit = 20000

var decoder = schema.NewDecoder()

// DB Entities

// Sitemap - an entity representing a particular sitemap
type Sitemap struct {
	ID   int    `schema:"-" storm:"id,increment"`
	Name string `schema:"name" storm:"unique"`
	URL  string `schema:"url" storm:"unique"`
}

type CrawlingHeader struct {
	Name  string
	Value string
}

// Crawling - a single crawling action, with page results related to it
type Crawling struct {
	ID          int              `schema:"-" storm:"id,increment"`
	SitemapID   int              `schema:"sitemap_id" storm:"index"`
	CreatedAt   time.Time        `schema:"-"`
	Processed   bool             `schema:"-" storm:"index" `
	Concurrency int              `schema:"sitemap_id"`
	Headers     []CrawlingHeader `schema:"headers"`
}

// PageResult - an entity representing a particular requested page
type PageResult struct {
	ID         int `storm:"id,increment"`
	CrawlingID int `storm:"index"`
	URL        string
	Status     int `storm:"index"`
}

// DB initialization

// NewDB - initializes a DB and ensures all the buckets are in place
// If resetData is true this will clean up database after initialization (used in tests).
func NewDB(filename string, resetData bool) *storm.DB {
	db, err := storm.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	if resetData {
		db.Drop(&Sitemap{})
		db.Drop(&Crawling{})
		db.Drop(&PageResult{})
	}
	db.Init(&Sitemap{})
	db.Init(&Crawling{})
	db.Init(&PageResult{})

	return db
}

// Actions
// Different functions for mutating data.

// CreateSitemap - creates a sitemap
func CreateSitemap(db *storm.DB, sitemap *Sitemap, r *http.Request) error {
	decoder.Decode(sitemap, r.Form)
	if len(sitemap.Name) == 0 || len(sitemap.URL) == 0 {
		return errors.New("Sitemap Name and URL cannot be empty")
	}
	return db.Save(sitemap)
}

// UpdateSitemap - updates a sitemap
func UpdateSitemap(db *storm.DB, sitemap *Sitemap, r *http.Request) error {
	decoder.Decode(sitemap, r.Form)
	if len(sitemap.Name) == 0 || len(sitemap.URL) == 0 {
		return errors.New("Sitemap Name and URL cannot be empty")
	}
	return db.Update(sitemap)
}

// DeleteSitemap - deletes a sitemap found by ID
func DeleteSitemap(db *storm.DB, id int) error {
	var sitemap Sitemap
	if err := db.One("ID", id, &sitemap); err != nil {
		return err
	}

	var crawlings []Crawling
	db.Find("SitemapID", sitemap.ID, &crawlings)

	tx, _ := db.Begin(true)
	tx.DeleteStruct(&sitemap)
	for _, crawling := range crawlings {
		var pageResults []PageResult
		db.Find("CrawlingID", crawling.ID, &pageResults)
		tx.DeleteStruct(&crawling)
		for _, pageResult := range pageResults {
			tx.DeleteStruct(&pageResult)
		}
	}
	return tx.Commit()
}

// CreateCrawling - creates a new crawling and prepares results for all
// the pages. Results are processed later in a background job.
// If there is a crawling already running for this sitemap, the new one
// will not be created and return will be returned instead.
func CreateCrawling(db *storm.DB, cr *Crawling, r *http.Request) error {
	sitemapID, _ := intValue(r, "sitemap_id")
	sitemap, err := GetSitemap(db, sitemapID)
	if err != nil {
		return err
	}

	query := db.Select(q.And(
		q.Eq("SitemapID", sitemapID),
		q.Eq("Processed", false),
	))

	if query.First(&Crawling{}) == nil {
		return errors.New("Current crawling is in progress, cannot create another")
	}

	urls, err := utils.ParseSitemap(sitemap.URL)
	if err != nil {
		return err
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	cr.CreatedAt = time.Now()
	decoder.Decode(cr, r.Form)
	tx.Save(cr)

	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range urls {
		tx.Save(&PageResult{
			CrawlingID: cr.ID,
			URL:        url,
			Status:     0,
		})
	}

	return tx.Commit()
}

// DeleteCrawling - deletes a crawling and all related page results
func DeleteCrawling(db *storm.DB, crawlingID int) error {
	var crawling Crawling

	if err := db.One("ID", crawlingID, &crawling); err != nil {
		return err
	}

	var pageResults []PageResult
	db.Find("CrawlingID", crawling.ID, &pageResults)

	tx, _ := db.Begin(true)
	tx.DeleteStruct(&crawling)
	for _, pageResult := range pageResults {
		tx.DeleteStruct(&pageResult)
	}
	return tx.Commit()
}

// Queries
// Basic query methods without any sophisticated logic

// GetSitemap - retrieves a sitemap by ID
func GetSitemap(db *storm.DB, sitemapID int) (*Sitemap, error) {
	sitemap := &Sitemap{}
	err := db.One("ID", sitemapID, sitemap)
	return sitemap, err
}

// GetSitemaps - retrieves all sitemaps for the store
func GetSitemaps(db *storm.DB) []Sitemap {
	var sitemaps []Sitemap
	db.All(&sitemaps)
	return sitemaps
}

// Filters
// For building advanced queries and using parameters

// CrawlingsFilter a struct used for filtering crawlings based on JSON encoded parameters
type CrawlingsFilter struct {
	Request *http.Request
}

// PageResultsFilter a struct used for filtering page results based on JSON encoded parameters
type PageResultsFilter struct {
	Request *http.Request
}

// Query - applies CrawlingsFilter and returns results
func (f *CrawlingsFilter) Query(db *storm.DB) []Crawling {
	crawlings := []Crawling{}
	filters := []q.Matcher{}

	if sitemapID, err := intValue(f.Request, "sitemap_id"); err == nil {
		filters = append(filters, q.Eq("SitemapID", sitemapID))
	}
	if processed, err := boolValue(f.Request, "processed"); err == nil {
		filters = append(filters, q.Eq("Processed", processed))
	}

	query := db.Select(q.And(filters...))
	query = query.Limit(limitValue(f.Request))
	query = query.Skip(offsetValue(f.Request))

	query.Find(&crawlings)
	return crawlings
}

// Query - applies PageResultsFilter and returns results
func (f *PageResultsFilter) Query(db *storm.DB) []PageResult {
	pageResults := []PageResult{}
	filters := []q.Matcher{}

	if crawlingID, err := intValue(f.Request, "crawling_id"); err == nil {
		filters = append(filters, q.Eq("CrawlingID", crawlingID))
	}
	if status, err := intValue(f.Request, "status"); err == nil {
		filters = append(filters, q.Eq("Status", status))
	}
	if url, err := stringValue(f.Request, "url"); err == nil {
		filters = append(filters, q.Re("URL", url))
	}

	query := db.Select(q.And(filters...))
	query = query.Limit(limitValue(f.Request))
	query = query.Skip(offsetValue(f.Request))

	query.Find(&pageResults)
	return pageResults
}

// Query helper functions
func stringValue(r *http.Request, name string) (string, error) {
	val := r.FormValue(name)
	if val == "" {
		return val, errors.New("is empty")
	}
	return val, nil
}

func intValue(r *http.Request, name string) (int, error) {
	val := r.FormValue(name)
	return strconv.Atoi(val)
}

func boolValue(r *http.Request, name string) (bool, error) {
	val := r.FormValue(name)

	if val == "true" {
		return true, nil
	} else if val == "false" {
		return false, nil
	}
	return false, errors.New("is empty")
}

func offsetValue(r *http.Request) int {
	offset, _ := intValue(r, "offset")
	return offset
}

func limitValue(r *http.Request) int {
	if limit, err := intValue(r, "limit"); err == nil {
		if limit > 0 && limit < queryLimit {
			return limit
		}
		return queryLimit
	}
	return queryLimit
}
