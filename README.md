<div align="center"><img src="https://github.com/luchkonikita/canary/blob/master/logo.png" width="800" /></div>

<br/>

![Travis Badge](https://travis-ci.org/luchkonikita/canary.svg?branch=master)

Canary is a microservice for testing your websites.
It can be deployed as a single binary and used via HTTP API.
You can add your sitemap and the application will request all the
pages and store results. With this done, you can detect any broken
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

To ping the API and get endpoints description you can just make a request with `curl localhost:4000`.

## Testing

Run `go test ./...` from the project root.

## TODO

- [ ] Store pages HTML
