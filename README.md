<p align="center">
    <img alt="Project Logo" src="./assets/logo/shorty.png"/>
</p>
<p align="center">
    <a alt="Build Status" href="https://travis-ci.com/otaviof/shorty">
        <img src="https://travis-ci.com/otaviof/shorty.svg?branch=master">
    </a>
    <a alt="Code Coverage" href="https://codecov.io/gh/otaviof/shorty">
        <img src="https://codecov.io/gh/otaviof/shorty/branch/master/graph/badge.svg">
    </a>
</p>

# `shorty`

Shorty is yet another URL shortener application. Basically, you inform a URL plus a arbitrary
string to Shorty, and then based on the short string Shorty will redirect your request to the
original URL. Simple like that.

# Running

Install `shorty` using `go get`:

``` bash
go get -x -u github.com/otaviof/shorty/cmd/shorty
```

And then:

``` bash
shorty --database-file /var/tmp/shorty.sqlite
```

## Docker

To run Shorty via Docker, use:

``` bash
docker run --publish "8000:8000" otaviof/shorty:latest
```

Alternatively, you can share a local volume to persist its database file:

``` bash
docker run --publish "8000:8000" --volume "<VOLUME_PATH>:/var/lib/shorty" otaviof/shorty:latest
```

# Usage

The following example shows how to add a short link to a URL via `curl`.

``` bash
curl -X POST http://127.0.0.1:8000/shorty -d '{ "url": "https://github.com/otaviof/shorty" }'
```

As output, you should see:

``` json
{
  "short": "shorty",
  "url": "https://github.com/otaviof/shorty",
  "created_at": 1553442790
}
```

And then, to `curl` with redirect to original URL:

``` bash
curl -L http://127.0.0.1:8000/shorty
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

Over the endpoint `/metrics` this application offers Prometheus compatible metrics. For instance:

``` bash
$ curl -s http://127.0.0.1:8000/metrics |tail
http_response_size_bytes{handler="read",quantile="0.5"} 0
http_response_size_bytes{handler="read",quantile="0.9"} 0
http_response_size_bytes{handler="read",quantile="0.99"} 0
http_response_size_bytes_sum{handler="read"} 0
http_response_size_bytes_count{handler="read"} 4
http_response_size_bytes{handler="slash",quantile="0.5"} 16
http_response_size_bytes{handler="slash",quantile="0.9"} 16
http_response_size_bytes{handler="slash",quantile="0.99"} 16
http_response_size_bytes_sum{handler="slash"} 32
http_response_size_bytes_count{handler="slash"} 2
```

The endpoints are named `read`, `create` and `slash`, where `read` redirect the requests while
`create` receive POST requests to save new URLs. The root endpoint, or `slash`, only display the
application name.

# Persistence

Backend storage is currently using SQLite. This application creates a `table` that's able to store
the records from the REST interface, and does not allow repetition of short strings.

On command-line or environment you can specify the location of the database file, by default data is
located on `/var/lib/shorty` directory.

# Development

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

### Unit-Tests

Unit tests are located on `pkg/shorty` directory, and using the suffix `_test`. To run unit-tests:

``` bash
make test
```

### Integration-Tests

Integration tests are on `test/e2e` directory. To run integration-tests:


``` bash
make integration
```
