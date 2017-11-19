package test_support

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func NewTestRequest(data map[string]string) *http.Request {
	json, _ := json.Marshal(data)
	// We do not care about routing details as this is used to test handlers directly
	return httptest.NewRequest("GET", "/", bytes.NewReader(json))
}

func NewServer(data string, header int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(header)
		fmt.Fprintln(w, data)
	}))
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
