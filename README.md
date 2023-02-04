# Transit CLI 🚇

`transit` is a cross platform tool that brings local transit information to the command line.

### Demo

https://user-images.githubusercontent.com/23173408/216787213-bf213727-73d7-4579-87c4-052333ceac87.mp4

## Installation

### Requirements

- An api key from your local transit authority

| location | website                      |
| -------- | ---------------------------- |
| dmv      | https://developer.wmata.com/ |

### Homebrew

```bash
# to install
brew tap ismailshak/tap
brew install transit

# to update to a newer version
brew update && brew upgrade transit
```

### Binaries download

Head over to [releases](https://github.com/ismailshak/transit/releases) for pre-built versions of `transit`. Find the executable for your system and download it.

### From source

Go will automatically install it in your `$GOPATH/bin` directory which should already be in your `$PATH`.

```bash
go install github.com/ismailshak/transit@latest
```

### Setup

For first time setup, you will need to run the following commands:

```bash
transit config set core.location <location>
transit config set <location>.api_key <api_key>

# example
transit config set core.location dmv
transit config set dmv.api_key abcd
```

## Usage

### Version

```bash
transit --version
```

### Help

A global flag that can be used with any command and subcommand. This will print usage with examples.

```bash
transit --help

transit <command> --help

transit <command> <subcommand> --help
```

### Verbose logging

A global flag that can be used with any command and subcommand. This will enable debug logs.

```bash
transit -v
transit --verbose
```

### Config

Interact with transit's configuration file.

Nested fields are addressable by using a period/full-stop/dot (`.`) as a delimiter e.g. `core.location`.

```bash
# print a config option
transit config get <key>

# set a config option
transit config set <key> <value>

# print path to config file
transit config path
```

<details>

<summary>Full list of available config options</summary>

```yaml
core:
  location: <string>

dmv:
  api_key: <string>
```

</details>

### List arriving trains for a station

```bash
# single station
transit list <station_name>

# multiple stations
transit list <station-1> <station-2> <station-3>

# examples
transit list --help
```

Arguments will be fuzzy matched against the list of official metro names, so it's a bit lenient. For example, "courthouse" will match "Court House" and "franconia" will match "Franconia-Springfield".

## Exit Codes

The following is a list of exit codes returned by `transit`. Status codes were referenced from the [OpenBSD source code](https://github.com/openbsd/src/blob/master/include/sysexits.h).

| Code | Explanation                  |
| ---- | ---------------------------- |
| 0    | Success                      |
| 64   | Incorrect usage of command   |
| 78   | Error due to bad config file |
