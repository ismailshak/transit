# Transit CLI 🚇

`transit` is a cross platform tool that brings local transit information to the command line.

### Demo

https://user-images.githubusercontent.com/23173408/216838894-066bbaa0-bfa9-4762-8f59-72aff4a735c5.mp4

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

This will print usage with examples. Help is a global flag that can be used with any command and subcommand.

```bash
transit -h
transit --help

transit <command> --help

transit <command> <subcommand> --help
```

### Verbose

This will enable debug logs. Verbose is a global flag that can be used with any command and subcommand.

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

<summary>Expand for full list of available config options</summary>

```yaml
core:
  location: <string>
  watch_interval: 10 # default

dmv:
  api_key: <string>
```

</details>

### At

Print upcoming trains at one or more stations.

```bash
# single station
transit at <station_name>

# multiple stations
transit at <station-1> <station-2> <station-3>

# watch mode, --watch or -w (one or many stations)
transit at <station> --watch

# examples
transit at --help
```

Arguments will be fuzzy matched against the list of official metro names, so it's a bit lenient. For example, "courthouse" will match "Court House" and "franconia" will match "Franconia-Springfield".

## Exit Codes

The following is a list of exit codes returned by `transit`. Status codes were referenced from the [OpenBSD source code](https://github.com/openbsd/src/blob/master/include/sysexits.h).

| Code | Explanation                  |
| ---- | ---------------------------- |
| 0    | Success                      |
| 64   | Incorrect usage of command   |
| 78   | Error due to bad config file |
