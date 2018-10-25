![Travis Badge](https://travis-ci.org/luchkonikita/canary.svg?branch=master)

Canary is a microservice for testing your websites.
It can be deployed as a single binary and used via HTTP API or a build-in web interface.
You can add your sitemap and the application will request all the pages and store results.
With this done, you can detect any broken resources on your website.

The service uses [Bolt](https://github.com/boltdb/bolt) as a storage and stores all
the data locally. You don't need any external store for making it work.

With the embedded storage Canary can provide you with some historical data if needed.

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
      origin to allow cross-origin requests (default "http://localhost:4000")
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

- GET /crawlings
	Returns a list of crawlings.

- POST /crawlings
	{
		"url": "http://example.com/sitemap.xml",
		"concurrency": 10
	}
	Creates a new crawling and starts it.

- GET /crawlings/1
	{}
	Returns a crawling for specified ID.

- DELETE /crawlings/1
	{}
	Deletes a crawling for specified ID.
```

Web interface is available on http://localhost:4000.

## Testing

Run `go test ./...` from the project root.

## TODO

- [x] When crawling is deleted need to turn off the WIP worker.
- [ ] Optimize front-end build.
- [x] Debounce form inputs.
- [x] Clipboard.
- [x] Add ETA to the crawling card.
- [x] Show report summary and details.
- [x] Handle the hanged worker when the connectivity goes down.
- [ ] Split `Crawling` UI component into smaller chunks.
- [x] Do better JSON serialization.
- [ ] Introduce timeout option for the crawling.
- [ ] Use `dep`.
- [ ] Store pages HTML.
- [ ] Store time for requests and show some statistics.
- [ ] Update README.
