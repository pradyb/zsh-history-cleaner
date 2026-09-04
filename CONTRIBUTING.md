# Contributing

## Setup

```sh
git clone https://github.com/pradyb/zsh-history-cleaner
cd zsh-history-cleaner
go build ./...
```

## Before opening a PR

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .   # must print nothing
```

CI runs the same checks on every push and PR.

## Guidelines

- Keep the tool dependency-free (stdlib only) unless there's a strong reason not to.
- Add a test for any new filter pass or parsing behaviour.
- Update the README's flags/features tables if you change user-facing behaviour.
