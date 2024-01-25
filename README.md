# eXtractor Tool

Extract Everything. Decompress entire folders filled with archives.

Simple command-line tool that works on mac, windows and linux.

## Installation

Use the directions below to install `xt` on your system.

### Linux

If you use **Alpine** or **Arch**, you can find a package for your OS on the [releases] page

Running the following script will install the GoLift package repo (using packagecloud.io).
This only works on Debian/Ubuntu and RedHat/Fedora Linux.
This is the recommend installation method for all users.

```shell
curl -s https://golift.io/repo.sh | sudo bash -s - xt
```

### macOS

- There's a binary on the [releases] page, but you should use Homebrew:

```shell
brew install golift/mugs/xt
```

### Windows

- Download an `exe` from the [releases] page and put it in your `PATH` somewhere.
- The tool only works in a command or terminal window.

### FreeBSD

- Download a FreeBSD binary from the [releases] page and extract it into your PATH somewhere.
- Recommend `/usr/local/bin`.
- _Maybe one day [GoReleaser] will support FreeBSD packages._

### Others

- This should work on any operating system that supports GoLang.
- After you [install go](https://go.dev/doc/install), install the `xt` app using `go install`:

```shell
cd /tmp
go get github.com/Unpackerr/xt
go install github.com/Unpackerr/xt
```

[releases]: https://github.com/Unpackerr/xt/releases
[GoReleaser]: https://github.com/goreleaser/goreleaser