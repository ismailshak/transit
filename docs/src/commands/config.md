# The `config` command

Interact with transit's configuration file. Check the [configuration reference](../config-reference.md) to see the available options.

Nested fields are addressable by using a period/full-stop/dot (`.`) as a delimiter e.g. `core.location`.

```
transit config <subcommand>
```

## Subcommands

- `get` - Print the current value of a config option

```
transit config get <key>
```

- `set` - Sets a config option

```
transit config set <key> <value>
```

- `path` - Prints the path to the config file

```
transit config path
```
