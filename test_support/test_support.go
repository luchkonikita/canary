package test_support

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const SitemapXML = `
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

func GetTestDBName() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename) + "/test_storage.db"
}

func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

func ContainsJSON(s string, key string, value string) bool {
	return strings.Contains(s, fmt.Sprintf("\"%s\":%s", key, value))
}
