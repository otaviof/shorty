<p align="center">
    <img alt="Project Logo" src="https://raw.githubusercontent.com/otaviof/shorty/master/assets/logo/shorty.png"/>
</p>
<p align="center">
    <a href="https://goreportcard.com/report/github.com/otaviof/shorty">
        <img src="https://goreportcard.com/badge/github.com/otaviof/shorty">
    </a>
    <a href="https://codecov.io/gh/otaviof/shorty">
        <img src="https://codecov.io/gh/otaviof/shorty/branch/master/graph/badge.svg">
    </a>
    <a href="https://pkg.go.dev/github.com/otaviof/shorty/pkg/shorty">
        <img src="https://img.shields.io/badge/pkg.go.dev-godoc-007d9c?logo=go&logoColor=white">
    </a>
    <a href="https://travis-ci.com/otaviof/shorty">
        <img src="https://travis-ci.com/otaviof/shorty.svg?branch=master">
    </a>
    <a href="https://hub.docker.com/r/otaviof/shorty">
        <img src="https://img.shields.io/docker/cloud/build/otaviof/shorty.svg">
    </a>
</p>

# `shorty`

Shorty is yet another URL shortener application. Basically, you inform a URL plus a arbitrary
string to Shorty, and then based on the short string Shorty will redirect your request to the
original URL. Simple like that.

# Running

Install `shorty` using `go get`:

```sh
go get -x -u github.com/otaviof/shorty/cmd/shorty
```

And then:

```sh
shorty --database-file /var/tmp/shorty.sqlite
```

## Docker

Images are stored on [Docker-Hub](https://hub.docker.com/r/otaviof/shorty), tags using a version
number are stable releases, and `master` is built from latest commits on branch.

For example, use:

```sh
docker run --publish "8000:8000" otaviof/shorty:latest
```

Alternatively, you can share a local volume to persist its database file:

```sh
docker run --publish "8000:8000" --volume "<VOLUME_PATH>:/var/lib/shorty" otaviof/shorty:latest
```

# Usage

The following example shows how to add a short link to a URL via `curl`.

```sh
curl -X POST http://127.0.0.1:8000/shorty/shorty -d '{ "url": "https://github.com/otaviof/shorty" }'
```

As output, you should see:

```json
{
  "short": "shorty",
  "url": "https://github.com/otaviof/shorty",
  "created_at": 1553442790
}
```

And then, to `curl` with redirect to original URL:

```sh
curl -L http://127.0.0.1:8000/shorty/shorty
```

## Command-Line Arguments

Application configuration can also be set via environment variables, or command-line parameters,
where the environment overwrites command-line. So for instance, if you want to set `--address`
option, you can export `SHORTY_ADDRESS` in environment. By setting the prefix as application name
(`SHORTY_`), followed by option name, in this case `ADDRESS`, split by underscore and all capitals.

The basic usage is:

``` bash
shorty [flags]
```

Where the following flags are available:

- `--address`: address and port to listen on;
- `--database-file`: database file path;
- `--idle-timeout`: idle connection timeout, in seconds;
- `--read-timeout`: read timeout, in second;
- `--write-timeout`: write timeout, in seconds;
- `--sqlite-flags`: connection string SQLite flags;
- `--help`: shows command-line help message;

## Instrumentation

Over the endpoint `/metrics` this application offers Prometheus compatible metrics, those are
collected using [OpenCensus](https://opencensus.io/):

```sh
$ curl -s http://127.0.0.1:8000/metrics |tail
opencensus_io_http_server_response_bytes_bucket{le="6.7108864e+07"} 4
opencensus_io_http_server_response_bytes_bucket{le="2.68435456e+08"} 4
opencensus_io_http_server_response_bytes_bucket{le="1.073741824e+09"} 4
opencensus_io_http_server_response_bytes_bucket{le="4.294967296e+09"} 4
opencensus_io_http_server_response_bytes_bucket{le="+Inf"} 4
opencensus_io_http_server_response_bytes_sum 3845
opencensus_io_http_server_response_bytes_count 4
# HELP opencensus_io_http_server_response_count_by_status_code Server response count by status code
# TYPE opencensus_io_http_server_response_count_by_status_code counter
opencensus_io_http_server_response_count_by_status_code{http_status="200"} 4
```

You can find documentation about HTTP metrics on OpenCensus
[documentation](https://opencensus.io/guides/http/go/net_http/server/#metrics). Furthermore, Shorty
is integrated with [OCSQL](https://github.com/opencensus-integrations/ocsql), you can read recorded
metrics [documentation here](https://github.com/opencensus-integrations/ocsql#recorded-metrics).

# Persistence

Backend storage is currently using SQLite. This application creates a `table` that's able to store
the records from the REST interface, and does not allow repetition of short strings.

On command-line or environment you can specify the location of the database file, by default data is
located on `/var/lib/shorty` directory.

# Contributing

## Project Structure

Following a description of directory and files important on how this project is organized.

| Folder       | Role  | Description                    |
|--------------|-------|--------------------------------|
| `assets`     | doc   | Contains project logo          |
| `cmd/shorty` | cmd   | Command line entrypoint        |
| `pkg/shorty` | pkg   | Shorty package                 |
| `test/e2e`   | tst   | Integration tests              |
| `vendor`     | build | Vendor directory, dependencies |

Regarding the relevant files:

| File              | Role    | Description                                        |
|-------------------|---------|----------------------------------------------------|
| `.goreleaser.yml` | build   | Build the project and organize a release on Github |
| `.travis.yml`     | CI      | Drives Travis-CI actions                           |
| `Dockerfile`      | build   | Docker image manifest                              |
| `Gopkg.*`         | build   | Dep, vendor management                             |
| `Makefile`        | build   | Automation of actions against project              |
| `version`         | version | Carry project version in a text file               |

## Testing

Unit tests are located on `pkg/shorty` directory, and using the suffix `_test`. To run unit-tests:

```sh
make test-unit
```

Integration tests are on `test/e2e` directory. To run integration-tests:


```sh
make test-e2e
```

And all tests can be triggered with:

```sh
make test
```
