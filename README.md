# Transit CLI 🚇

`transit` is a cross platform tool that brings local transit information to the command line.

### Demo

[![asciicast](https://asciinema.org/a/Qq39eZP80bsdNLwcW5JCj1IzT.svg)](https://asciinema.org/a/Qq39eZP80bsdNLwcW5JCj1IzT)

## Installation

### Requirements

- An api key for you local transit authority.

| location | website                      |
| -------- | ---------------------------- |
| dmv      | https://developer.wmata.com/ |

### Binary

Head over to [releases](https://github.com/ismailshak/transit/releases), find the executable for your system and download it.

### Homebrew

```bash
brew install ismailshak/tap/transit
```

## Usage

### List arriving trains for a station

You can provide 1 or more arguments to `list`. Arguments will be fuzzy matched against the list of official metro names, so it's a bit lenient. For example, "courthouse" will match "Court House" and "franconia" will match "Franconia-Springfield".

If a provided argument is too generic, it will be skipped. Try narrowing your argument so that it only matches 1 station name.

```bash
# single station
transit list <station_name>
transit list rosslyn
transit list "u street"
transit list "metro center"

# multiple stations
transit list <station-1> <station-2> <station-3>
transit list courthouse rosslyn "u street"
```

### Help messages

```shell
transit --help
transit <subcommand> --help
```

### Version

```shell
transit --version
```
