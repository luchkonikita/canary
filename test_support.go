package main

// TestSitemapXML - a fixture containing sitemap XML.
const TestSitemapXML = `
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	  <url>
		<loc>http://google.com/maps</loc>
		<lastmod>2016-04-04T02:08:53+03:00</lastmod>
		<priority>1.000000</priority>
	  </url>
	  <url>
		<loc>http://google.com/docs</loc>
		<lastmod>2016-04-04T01:12:13+03:00</lastmod>
		<priority>1.000000</priority>
	  </url>
	</urlset>
`

func mockApp() *application {
	app := newApplication("test_storage.db")
	app.db.Drop(&crawling{})
	return app
}
