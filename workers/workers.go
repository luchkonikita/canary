package workers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/luchkonikita/canary/store"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

const (
	concurrency = 7
)

var (
	workTimeout  = time.Second * 10
	workRestarts = int(math.Pow(10, 6))
)

// CrawlerWorker - the worker to run crawlings
type CrawlerWorker struct {
	client   *http.Client
	sitemap  store.Sitemap
	crawling store.Crawling
	db       *storm.DB
}

func (cw *CrawlerWorker) String() string {
	return fmt.Sprintf("[Crawler worker %d]", cw.sitemap.ID)
}

// Start - starts a pool of workers and updates it when new sitemaps are added.
func Start(db *storm.DB) {
	pool := make(map[int]bool)

	for {
		log.Println("[Main jobs thread]: Loading sitemaps")
		var sitemaps []store.Sitemap
		db.All(&sitemaps)

		for _, sitemap := range sitemaps {
			if pool[sitemap.ID] {
				continue
			} else {
				cw := NewCrawlerWorker(db, sitemap)
				pool[sitemap.ID] = true
				log.Printf("%s: Worker started \n", cw)
				go cw.Work(workRestarts)

			}
		}

		// Wait and query again in case new sitemaps were added.
		time.Sleep(time.Minute)
	}
}

// NewCrawlerWorker - returns new CrawlerWorker
func NewCrawlerWorker(db *storm.DB, sitemap store.Sitemap) *CrawlerWorker {
	return &CrawlerWorker{
		db:      db,
		sitemap: sitemap,
		client:  &http.Client{},
	}
}

// Work - retrieves scheduled crawlings and requests all the related URLs
// For each request the result is stored in the corresponding PageResult
// If the crawling is deleted via API the worker will stop the execution
//
// Note the `restarts` parameter, providing a flexible API.
// When called with default value int(math.Pow(10, 6)) the worker will run or ~115 days.
// With customizable restarts this function is easier to test.
func (cw *CrawlerWorker) Work(restarts int) {
	db := cw.db

	restartIfNeeded := func() {
		if restarts > 0 {
			time.Sleep(workTimeout)
			cw.Work(restarts - 1)
		}
	}

	query := db.Select(
		q.And(
			q.Eq("SitemapID", cw.sitemap.ID),
			q.Eq("Processed", false),
		),
	)
	err := query.First(&cw.crawling)

	if err == storm.ErrNotFound {
		restartIfNeeded()
		return
	}

	log.Printf("%s: Crawling started \n", cw)

	query = db.Select(
		q.And(
			q.Eq("CrawlingID", cw.crawling.ID),
			q.Eq("Status", 0),
		),
	)

	// TODO: Move concurrency to crawling
	var concurrency int
	if cw.crawling.Concurrency > 0 && cw.crawling.Concurrency < 20 {
		concurrency = cw.crawling.Concurrency
	} else {
		concurrency = 1
	}

Loop:
	for {
		var pageResults []store.PageResult

		// The query is executed in a loop fetching a limited number of records.
		// Each loaded batch is processed concurrently.
		// When there are no batches left, crawling is marked as completed.
		err := query.Limit(concurrency).Find(&pageResults)

		if err != nil {

			// TODO: Re-implement this function
			markCrawlingAsDone(db, &cw.crawling)
			break Loop
		} else {
			cw.processPageResults(pageResults)
		}
	}

	log.Printf("%s: Crawling completed \n", cw)
	restartIfNeeded()
}

// When crawling has 0 pending page  results it is marked as done and won't
// be processed in the future.
func markCrawlingAsDone(db *storm.DB, crawling *store.Crawling) {
	log.Printf("Completed crawling %d", crawling.ID)
	crawling.Processed = true
	err := db.Update(crawling)
	if err != nil {
		log.Println(err)
	}
}

// Wrapper around processPageResult to make requests in batches.
// Spawns each request in a separate goroutine and blocks until all requests resolve.
func (cw *CrawlerWorker) processPageResults(pageResults []store.PageResult) {
	var wg sync.WaitGroup

	for _, pageResult := range pageResults {
		wg.Add(1)

		go func(pageResult store.PageResult) {
			defer wg.Done()
			err := cw.processPageResult(pageResult)
			if err != nil {
				log.Println(err)
			}
		}(pageResult)
	}
	wg.Wait()
}

func (cw *CrawlerWorker) processPageResult(pageResult store.PageResult) error {
	req, err := http.NewRequest("GET", pageResult.URL, nil)
	if err != nil {
		return err
	}

	for _, header := range cw.crawling.Headers {
		req.Header.Add(header.Name, header.Value)
	}

	res, err := cw.client.Do(req)
	if err != nil {
		return err
	}

	pageResult.Status = res.StatusCode
	return cw.db.Update(&pageResult)
}
