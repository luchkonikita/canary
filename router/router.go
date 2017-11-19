package router

import (
	"github.com/luchkonikita/canary/handlers"

	"github.com/asdine/storm"
	"github.com/julienschmidt/httprouter"
)

// NewRouter - defines all the routes and corresponding handlers
func NewRouter(db *storm.DB) *httprouter.Router {
	router := httprouter.New()

	// Sitemaps
	router.GET("/sitemaps", handlers.GetSitemaps(db))
	router.POST("/sitemaps", handlers.CreateSitemap(db))
	router.DELETE("/sitemaps/:sitemapId", handlers.DeleteSitemap(db))

	// Crawlings
	router.GET("/sitemaps/:sitemapId/crawlings", handlers.GetCrawlings(db))
	router.POST("/sitemaps/:sitemapId/crawlings", handlers.CreateCrawling(db))
	router.DELETE("/sitemaps/:sitemapId/crawlings/:crawlingId", handlers.DeleteCrawling(db))

	// Page results
	router.GET("/crawlings/:crawlingId/page_results", handlers.GetPageResults(db))

	return router
}
