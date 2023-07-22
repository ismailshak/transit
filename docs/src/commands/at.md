# The `at` command

Display upcoming train arrival information to the terminal:

```bash
transit at <station>
```

You can provide one or multiple arguments and you don't have to type the full station name.

Arguments are fuzzy matched against the list of available stations and an argument is accepted only if a single station can be matched against it. For example, if you want to display information for the Franconia-Springfield station you can enter something like; "francon", "springfi" or even "fransp". If an argument matched too many stations (or none), it will be skipped.

## Options

- `-w`, `--watch`

When you use the `--watch` (`-w`) flag, the displayed output will automatically refresh after a certain amount of time. The refresh intervel is determined by the value of `watch_interval` in your config.
