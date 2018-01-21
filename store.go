package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

// DB Entities

type RequestHeader struct {
	Name  string
	Value string
}

// Crawling - a single crawling action, with page results related to it
type Crawling struct {
	ID          int             `schema:"-" storm:"id,increment"`
	CreatedAt   time.Time       `schema:"-"`
	URL         string          `schema:"url" storm:"index"`
	Processed   bool            `schema:"-" storm:"index" `
	Concurrency int             `schema:"concurrency"`
	Headers     []RequestHeader `schema:"headers"`
}

// PageResult - an entity representing a particular requested page
type PageResult struct {
	ID         int `storm:"id,increment"`
	CrawlingID int `storm:"index"`
	URL        string
	Status     int `storm:"index"`
}

// Validate - return an error if crawling is invalid
func (cr *Crawling) Validate() error {
	// TODO: Use regexp to match URL structure
	if len(cr.URL) == 0 {
		return errors.New("Crawling should have a valid URL")
	}
	if cr.Concurrency == 0 {
		return errors.New("Crawling should have a concurrency bigger than 0")
	}
	return nil
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
		db.Drop(&Crawling{})
		db.Drop(&PageResult{})
	}
	db.Init(&Crawling{})
	db.Init(&PageResult{})

	return db
}

// Actions
// Different functions for mutating data.

// CreateCrawling - creates a new crawling and prepares results for all
// the pages. Results are processed later in a background job.
// If there is a crawling already running for this sitemap, the new one
// will not be created and return will be returned instead.
func CreateCrawling(db *storm.DB, cr *Crawling, r *http.Request) error {
	cr.CreatedAt = time.Now()
	decoder.Decode(cr, r.Form)

	if err := cr.Validate(); err != nil {
		return err
	}

	query := db.Select(q.And(
		q.Eq("URL", cr.URL),
		q.Eq("Processed", false),
	))

	if query.First(&Crawling{}) == nil {
		return errors.New("Current crawling is in progress, cannot create another")
	}

	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	urls, err := ParseSitemap(cr)
	if err != nil {
		return err
	}

	tx.Save(cr)
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
