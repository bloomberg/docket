# dkt

Command `dkt` runs `docker-compose` with a set of docket files.

## Usage

`dkt` is useful for working with your docket configurations without having to
invoke `go test`.

Running `dkt -h` will show `dkt`'s help followed by `docker-compose`'s help.

```console
$ dkt -h

dkt runs docker-compose with the docker-compose files and generated
configuration that match the given docket mode and prefix.

Any arguments that aren't dkt-specific will be passed through to docker-compose.

Usage:
  dkt [OPTIONS] [arguments to docker-compose...]

Examples:
  dkt config
  dkt up -d
  dkt down

Options:
  -h, --help            Show this help
  -v, --version         Show version information
  -m, --mode=MODE       Set the docket mode (required) [$DOCKET_MODE]
  -P, --prefix=PREFIX   Set the docket prefix (default: docket) [$DOCKET_PREFIX]

Output of 'docker-compose help'
-------------------------------
...
```

### Example

While working on a feature, you might want to run a particular docket-based
test(s) in a tight loop.

```sh
DOCKET_MODE=mode go test -run testPattern
```

By leaving out `DOCKET_DOWN=1`, the Docker Compose app will stay up between each
run of `go test`, making the tests start more quickly.

When you're done testing, you'll want to shut down the Docker Compose app. You
_could_ do this by running the test against and adding `DOCKET_DOWN=1`, but that
means waiting while the test(s) run again. Instead, you can use `dkt down` to
run `docker-compose down`.

```sh
DOCKET_MODE=mode dkt down
# or
dkt -m mode down
```

## Installation

We highly recommend building `dkt` in module-mode. To do this, you can use a
tool like [`gobin`](https://github.com/myitcv/gobin) or do it yourself in a
temporary directory like so:

```sh
dktdir=$(mktemp -d)
cd "$dktdir"

go mod init dktmod # make up any name you like

go install github.com/bloomberg/docket/dkt

cd
rm -rf "$dktdir"
```

If you have docket inside your `$GOPATH/src` tree, you can run it on the fly
using `go run`:

```sh
go run github.com/bloomberg/docket/dkt
```
