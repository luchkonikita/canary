package store

import (
	"errors"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

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

// DeleteSitemap - deletes a sitemap found by ID
func DeleteSitemap(db *storm.DB, id int) error {
	// TODO: This should delete all related entities
	var sitemap Sitemap
	db.One("ID", id, &sitemap)
	return db.DeleteStruct(&sitemap)
}

// GetCrawlings - retrieves all the crawling for a particular sitemap
func GetCrawlings(db *storm.DB, sitemap *Sitemap) []Crawling {
	crawlings := []Crawling{}
	db.Find("SitemapID", sitemap.ID, &crawlings)
	return crawlings
}

// CreateCrawling - creates a new crawling and prepares results for all
// the pages. Results are processed later in a background job.
// If there is a crawling already running for this sitemap, the new one
// will not be created and return will be returned instead.
func CreateCrawling(db *storm.DB, sitemap *Sitemap, urls []string) error {
	query := db.Select(q.And(
		q.Eq("SitemapID", sitemap.ID),
		q.Eq("Processed", false),
	))
	err := query.First(&Crawling{})

	if err == nil {
		return errors.New("Current crawling is in progress, cannot create another")
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	currentTime := time.Now()

	crawling := &Crawling{
		SitemapID: sitemap.ID,
		CreatedAt: currentTime,
		Processed: false,
	}
	tx.Save(crawling)

	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, url := range urls {
		tx.Save(&PageResult{
			CrawlingID: crawling.ID,
			URL:        url,
			Status:     0,
		})
	}

	return tx.Commit()
}

func DeleteCrawling(db *storm.DB, sitemapID int, crawlingID int) error {
	var crawling Crawling

	query := db.Select(
		q.And(
			q.Eq("SitemapID", sitemapID),
			q.Eq("ID", crawlingID),
		),
	)

	err := query.First(&crawling)
	if err != nil {
		return err
	}

	var pageResults []PageResult
	db.Find("CrawlingID", crawling.ID, &pageResults)

	tx, err := db.Begin(true)

	tx.DeleteStruct(&crawling)

	for _, pageResult := range pageResults {
		tx.DeleteStruct(&pageResult)
	}

	return tx.Commit()
}

func GetPageResults(db *storm.DB, crawling Crawling) []PageResult {
	var pageResults []PageResult
	db.Find("CrawlingID", crawling.ID, &pageResults)
	return pageResults
}
