package store

import (
	"errors"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

const queryLimit = 10000

// Sitemap - an entity representing a particular sitemap
type Sitemap struct {
	ID       int    `storm:"id,increment"`
	Name     string `storm:"unique"`
	URL      string `storm:"unique"`
	Username string
	Password string
}

// Crawling - a single crawling action, with page results related to it
type Crawling struct {
	ID        int `storm:"id,increment"`
	SitemapID int `storm:"index"`
	CreatedAt time.Time
	Processed bool `storm:"index"`
}

// PageResult - an entity representing a particular requested page
type PageResult struct {
	ID         int `storm:"id,increment"`
	CrawlingID int `storm:"index"`
	URL        string
	Status     int `storm:"index"`
}

// HasAuth - shows if sitemap has basic auth credentials
func (s Sitemap) HasAuth() bool {
	return len(s.Username) > 0 && len(s.Password) > 0
}

// NewDB - initializes a DB and ensures all the buckets are in place
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

func CreateSitemap(db *storm.DB, sitemap *Sitemap) error {
	if len(sitemap.Name) == 0 || len(sitemap.URL) == 0 {
		return errors.New("Sitemap Name and URL cannot be empty")
	}
	return db.Save(sitemap)
}

// UpdateSitemap - updates a sitemap
func UpdateSitemap(db *storm.DB, sitemap *Sitemap) error {
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
func CreateCrawling(db *storm.DB, sitemap *Sitemap, urls []string) (*Crawling, error) {
	crawling := &Crawling{}
	query := db.Select(q.And(
		q.Eq("SitemapID", sitemap.ID),
		q.Eq("Processed", false),
	))
	err := query.First(&Crawling{})

	if err == nil {
		return crawling, errors.New("Current crawling is in progress, cannot create another")
	}

	tx, err := db.Begin(true)
	if err != nil {
		return crawling, err
	}

	crawling.SitemapID = sitemap.ID
	crawling.CreatedAt = time.Now()
	crawling.Processed = false

	tx.Save(crawling)

	if err != nil {
		return crawling, err
	}
	defer tx.Rollback()

	for _, url := range urls {
		tx.Save(&PageResult{
			CrawlingID: crawling.ID,
			URL:        url,
			Status:     0,
		})
	}

	return crawling, tx.Commit()
}

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

// CrawlingsFilter a struct used for filtering crawlings based on JSON encoded parameters
type CrawlingsFilter struct {
	SitemapID int    `json:"sitemap_id"`
	Processed string `json:"processed,omitempty"`
	Limit     int    `json:"limit"`
}

// Query - applies CrawlingsFilter and returns results
func (f *CrawlingsFilter) Query(db *storm.DB) []Crawling {
	crawlings := []Crawling{}
	filters := []q.Matcher{}

	if f.SitemapID != 0 {
		filters = append(filters, q.Eq("SitemapID", f.SitemapID))
	}
	if f.Processed == "true" {
		filters = append(filters, q.Eq("Processed", true))
	}
	if f.Processed == "false" {
		filters = append(filters, q.Eq("Processed", false))
	}

	query := db.Select(q.And(filters...))

	if f.Limit > 0 && f.Limit < queryLimit {
		query = query.Limit(f.Limit)
	} else {
		query = query.Limit(queryLimit)
	}

	query.Find(&crawlings)
	return crawlings
}

// PageResultsFilter a struct used for filtering page results based on JSON encoded parameters
type PageResultsFilter struct {
	CrawlingID int    `json:"crawling_id"`
	URL        string `json:"url,omitempty"`
	Status     int    `json:"status,omitempty"`
	Limit      int    `json:"limit"`
}

// Query - applies PageResultsFilter and returns results
func (f *PageResultsFilter) Query(db *storm.DB) []PageResult {
	pageResults := []PageResult{}
	filters := []q.Matcher{}

	if f.CrawlingID != 0 {
		filters = append(filters, q.Eq("CrawlingID", f.CrawlingID))
	}
	if f.Status != 0 {
		filters = append(filters, q.Eq("Status", f.Status))
	}
	if f.URL != "" {
		filters = append(filters, q.Re("URL", f.URL))
	}

	query := db.Select(q.And(filters...))

	if f.Limit > 0 && f.Limit < queryLimit {
		query = query.Limit(f.Limit)
	} else {
		query = query.Limit(queryLimit)
	}

	query.Find(&pageResults)
	return pageResults
}
