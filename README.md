# Transit CLI 🚇

`transit` is a cross platform tool that brings local transit information to the command line.

### Demo

[![asciicast](https://asciinema.org/a/Qq39eZP80bsdNLwcW5JCj1IzT.svg)](https://asciinema.org/a/Qq39eZP80bsdNLwcW5JCj1IzT)

## Installation

### Requirements

- An api key for you local transit authority.
- Setup your config file following instructions in Usage > Config below

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

### Config

A config file will automatically be generated the first time you run `transit`.

For first time setup, you will need to run these commands:

```bash
transit config set core.location="<location>"
transit config set <location>.api_key="<api_key>"

# Examples
transit config set core.location="dmv"
transit config set dmv.api_key="abcd"
```

`config` has other commands:

```bash
# help
transit config --help

# print a config option
transit config get <key>

# print path to config file
transit config path
```

<details>

<summary>Full list of available config options</summary>

```
core.location

dmv.api_key
```

</details>

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
