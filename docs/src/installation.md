# Installation

There are multiple ways to install the transit CLI tool. Choose any one of the methods below that best suit your needs.

## Homebrew

Supports both macOS and Linux homebrew.

To install:

```bash
brew tap ismailshak/tap
brew install transit
```

To update to a newer version:

```bash
brew update && brew upgrade transit
```

## Pre-compiled binaries

Executable binaries are available for download on the [GitHub Releases](https://github.com/ismailshak/transit/releases) page. Download the binary for your platform (Windows, macOS, or Linux) and extract the archive. The archive contains the `transit` executable.

To make it easier to run, put the path to the binary into your `PATH`.

## Building from source using go

Go will automatically install it in your `$GOPATH/bin` directory which should already be in your `$PATH`.

```bash
go install github.com/ismailshak/transit@latest
```
