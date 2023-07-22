# Global Flags

The following flags are global, and can be passed to any transit command.

## Help

`-h`, `--help`. Print usage with examples

```bash
transit -h
transit --help

transit <command> --help

transit <command> <subcommand> --help
```

## Config

`-c`, `--config`. Override the default config file with a path to another.

```bash
transit -c <path>
transit --config <path>

transit <command> --config <path>

transit <command> <subcommand> --config <path>
```

## Verbose

`-v`, `--verbose`. This will enable debug logs.

```bash
transit -v
transit --verbose

transit <command> --verbose

transit <command> <subcommand> --verbose
```
