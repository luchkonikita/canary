<div align="center"><img src="https://github.com/luchkonikita/canary/blob/master/logo.png" width="800" /></div>

<br/>

![Travis Badge](https://travis-ci.org/luchkonikita/canary.svg?branch=master)

Canary is a microservice for testing your websites.
It can be deployed as a single binary and used via HTTP API.
You can add your sitemap and the application will request all the
pages and store resul With this done, you can detect any broken
resources on your website.

The service uses [Bolt](https://github.com/boltdb/bolt) as a storage and stores all
the data locally. You don't need any external store for making it work.

With the embedded storage Canary can provide you with historical data.

## Installation

```
go get github.com/luchkonikita/canary
go install github.com/luchkonikita/canary
```

## Usage

Use the `canary` from the command-line with the following flags:

```
-db string
      database file (default "canary.db")
-origin string
      origin to allow cross-origin requests (default "http://localhost:8080")
-password string
      password for basic auth (if needed)
-port string
      port to listen on (default "4000")
-username string
      username for basic auth (if needed)
```

This is the API shape.

```
Usage examples:

- GET /
      {}
      Pings the server.

- GET /crawlings
	{
		"url": "http://example.com/sitemap.xml",
		"processed": "true",
		"limit": 10,
		"offset": 0,
	}
	Returns a list of crawlings.

- POST /crawlings
	{
		"url": "http://example.com/sitemap.xml",
		"concurrency": 10
	}
	Creates a new crawling and starts it.

- DELETE /crawlings/1
	{}
	Deletes a crawling and cancels it.

- GET /page_results
	{
		"crawling_id": 1,
		"status": 500,
		"url": "some-url-substring"
		"limit": 10,
		"offset": 0,
	}
	Returns a list of page results.
```

## Testing

Run `go test ./...` from the project root.

## TODO

- [ ] Store pages HTML
