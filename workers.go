package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/asdine/storm/q"
	"github.com/remeh/sizedwaitgroup"
)

var (
	client       = &http.Client{}
	workTimeout  = time.Second * 10
	workRestarts = int(math.Pow(10, 6))
)

func (app *application) runWorkers() {
	pool := make(map[int]bool)

	for {
		log.Println("[Main jobs thread]: Loading pending crawlings")
		var crs []crawling
		query := app.db.Select(q.Eq("Processed", false))
		query.Find(&crs)

		for _, cr := range crs {
			if pool[cr.ID] {
				continue
			} else {
				pool[cr.ID] = true
				go app.processCrawling(cr)
			}
		}

		// Wait and query again in case new crawlings were added.
		time.Sleep(time.Minute)
	}
}

func (app *application) processCrawling(cr crawling) {
	log.Printf("[Crawler worker %d]: Crawling started \n", cr.ID)

loadLoop:
	for {
		allProcessed := true
		crawlingExists := true
		swg := sizedwaitgroup.New(cr.Concurrency)

		for i := range cr.PageResults {
			swg.Add()

			go func(index int) {
				defer swg.Done()
				if cr.PageResults[index].Status == 0 {
					status, err := requestPage(cr.PageResults[index].URL, cr.Headers)
					if err != nil {
						fmt.Println(err)
						allProcessed = false
					} else {
						cr.PageResults[index].Status = status
						err := app.db.Update(&cr)

						if err != nil {
							crawlingExists = false
						}
					}
				}
			}(i)
		}

		swg.Wait()

		// Crawling was deleted before finishing.
		if !crawlingExists {
			log.Printf("[Crawler worker %d]: Crawling interrupted \n", cr.ID)
			break loadLoop
		}

		// Crawling finished successfully.
		if allProcessed {
			cr.Processed = true
			app.db.Update(&cr)
			log.Printf("[Crawler worker %d]: Crawling completed \n", cr.ID)
			break loadLoop
		}
	}
}
