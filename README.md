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

There are no releases with a ready binary yet,
as the project is in active development and is not stable.

If you want to build it locally just clone the repo and do `go build` from the root.

## Usage

_Coming soon..._

## Testing

Run `go test ./...` from the project root.

## TODO

- [ ] Improve README
- [ ] Add skip option for the pagination
- [ ] Document endpoints
- [ ] Store pages HTML
- [ ] Configure concurrency in the sitemap record
