package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

var (
	client       = &http.Client{}
	workTimeout  = time.Second * 10
	workRestarts = int(math.Pow(10, 6))
)

// CrawlerWorker - the worker to run crawlings
type CrawlerWorker struct {
	crawling Crawling
	db       *storm.DB
}

func (cw *CrawlerWorker) String() string {
	return fmt.Sprintf("[Crawler worker %d]", cw.crawling.ID)
}

// StartWorkers - starts a pool of workers and updates it when new crawlings are started.
func StartWorkers(db *storm.DB) {
	pool := make(map[int]bool)

	for {
		log.Println("[Main jobs thread]: Loading pending crawlings")
		var crawlings []Crawling
		query := db.Select(q.Eq("Processed", false))
		query.Find(&crawlings)

		for _, crawling := range crawlings {
			if pool[crawling.ID] {
				continue
			} else {
				cw := &CrawlerWorker{
					db:       db,
					crawling: crawling,
				}
				pool[crawling.ID] = true
				go cw.Work()
			}
		}

		// Wait and query again in case new crawlings were added.
		time.Sleep(time.Minute)
	}
}

// Work - retrieves scheduled crawlings and requests all the related URLs
// For each request the result is stored in the corresponding PageResult
// If the crawling is deleted via API the worker will stop the execution
//
// Note the `restarts` parameter, providing a flexible API.
// When called with default value int(math.Pow(10, 6)) the worker will run or ~115 days.
// With customizable restarts this function is easier to test.
func (cw *CrawlerWorker) Work() {
	db := cw.db

	log.Printf("%s: Crawling started \n", cw)

	query := db.Select(
		q.And(
			q.Eq("CrawlingID", cw.crawling.ID),
			q.Eq("Status", 0),
		),
	)

Loop:
	for {
		var pageResults []PageResult

		// The query is executed in a loop fetching a limited number of records.
		// Each loaded batch is processed concurrently.
		// When there are no batches left, crawling is marked as completed.
		err := query.Limit(cw.crawling.Concurrency).Find(&pageResults)

		if err != nil {
			cw.markCrawlingAsDone()
			break Loop
		} else {
			cw.processPageResults(pageResults)
		}
	}

	log.Printf("%s: Crawling completed \n", cw)
}

// When crawling has 0 pending page results it is marked as done and won't
// be processed in the future.
func (cw *CrawlerWorker) markCrawlingAsDone() {
	crawling := cw.crawling
	crawling.Processed = true
	err := cw.db.Update(&crawling)
	if err != nil {
		log.Println(err)
	}
}

// Wrapper around processPageResult to make requests in batches.
// Spawns each request in a separate goroutine and blocks until all requests resolve.
func (cw *CrawlerWorker) processPageResults(pageResults []PageResult) {
	var wg sync.WaitGroup

	for _, pageResult := range pageResults {
		wg.Add(1)

		go func(pageResult PageResult) {
			defer wg.Done()
			err := cw.processPageResult(pageResult)
			if err != nil {
				log.Println(err)
			}
		}(pageResult)
	}
	wg.Wait()
}

func (cw *CrawlerWorker) processPageResult(pageResult PageResult) error {
	req, err := http.NewRequest("GET", pageResult.URL, nil)
	if err != nil {
		return err
	}

	for _, header := range cw.crawling.Headers {
		req.Header.Add(header.Name, header.Value)
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	pageResult.Status = res.StatusCode
	return cw.db.Update(&pageResult)
}
