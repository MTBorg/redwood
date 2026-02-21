# Development

## Project structure

This project follows [golang-standards
project-layout](https://github.com/golang-standards/project-layout).

```markdown
cmd/
    redwood/
        main.go # main binary
internal/
    # internal packages
```

## CLI

We use [Cobra](https://cobra.dev) for the CLI interface.

## Logging

This project uses the go standard library's `slog` package for logging.
By default logs are written to the logfile `/var/log/redwood.log`, use the
`--log-file` flag to change the path or `--log-file=stderr` to write logs to
stderr instead.
