# Canary

![Travis Badge](https://travis-ci.org/luchkonikita/canary.svg?branch=master)

<img src="https://cl.ly/2u0k2O2e233s/download/canary.jpg" width="200" />

Canary is a microservice for testing your websites.
It can be deployed as a single binary and used via HTTP API.
You can add your sitemap and the application will request all the
pages and store results. With this done, you can detect any broken
resources on your website.

The service uses [Bolt](https://github.com/boltdb/bolt) as a storage and stores all
the data locally. You don't need any external store for making it work.

With the embedded storage Canary can provide you with historical data.

_The more detailed decription is coming soon..._

## Installation

_Coming soon..._

## Usage

_Coming soon..._

## Testing

Run `go test ./...` from the project root.

## TODO

- [ ] Add pagination to page results endpoint
- [ ] Document endpoints
- [ ] Backup endpoint
- [ ] Data reset endpoint
- [ ] Store pages HTML
