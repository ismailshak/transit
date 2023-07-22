# Setup

Once you have the transit CLI tool installed, you're going to need to configure it.

## Api key

An api key from your local transit authority is required. Refer to the table below for the list of supported locations and where to go to generate a key.

| location | website                      |
| -------- | ---------------------------- |
| dmv      | https://developer.wmata.com/ |

## Config

For first time setup, you will need to run the following config commands after you've generated an api key.

### Set your location

```
transit config set core.location <location>
```

e.g. `transit config set core.location dmv`

### Set your api key

```
transit config set <location>.api_key <api_key>
```

e.g. `transit config set dmv.api_key abcd`
