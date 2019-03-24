<p align="center"><img src="./assets/logo/shorty.png"/></p>

# `shorty`

Shorty is yet another URL shortener service. Basically, you an informa a arbritary string to Shorty
plus a long URL, and next time you query the short string it will redirect the request to the
original URL.

# Installing

Via `go get`:

``` bash
go get -x -u github.com/otaviof/shorty/cmd/shorty
```

## Docker

To run Shorty via Docker, use:

``` bash
docker run --publish "8000:8000" otaviof/shorty:latest
```

Alternatively, you can share a local volume to persist its database, via:

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
option, you can export `SHORTY_ADDRESS`, having the prefix as application name  followed by option
name, split by underscore.

The following options are available:

- `--address`: address and port to listen on;
- `--database-file`: database file path;
- `--idle-timeout`: idle connection timeout, in seconds;
- `--read-timeout`: read timeout, in second;
- `--write-timeout`: write timeout, in seconds;
- `--sqlite-flags`: connection string SQLite flags;
- `--help`: to consider command-line help and now more about parameters;

# Persistence

Backend storage is currently using SQLlite. This application creates a table that's able to store
the records from the REST interface, and does not allow repetition of short strings.

# Development

## Project Structure

The most relevant directories are organized this way:

| Folder       | Role  | Description                    |
|--------------|-------|--------------------------------|
| `assets`     | doc   | Contains project logo          |
| `cmd/shorty` | cmd   | Command line entrypoint        |
| `pkg/shorty` | pkg   | Shorty package                 |
| `test/e2e`   | tst   | Integration tests              |
| `vendor`     | build | Vendor directory, dependencies |

And the most relevant files:

| File              | Role    | Description                                        |
|-------------------|---------|----------------------------------------------------|
| `.goreleaser.yml` | build   | Build the project and organize a release on Github |
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
