# `dkt`

Command `dkt` runs `docker-compose` with a set of docket files.

## Using `dkt`

`dkt` is useful for working with your docket setups without having to invoke
`go test`.

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

A common pattern is to install Go tools into `$GOPATH/bin`.

```sh
go get github.com/bloomberg/docket/dkt
```

If you have docket inside your `$GOPATH/src` tree, you can run it on the fly
using `go run`:

```sh
go run github.com/bloomberg/docket/dkt
```

## Usage

Run `dkt -h` for usage.
