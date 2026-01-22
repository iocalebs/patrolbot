# PatrolBot

[![codecov](https://codecov.io/gh/iocalebs/patrolbot/graph/badge.svg?token=2T23W9WZUL)](https://codecov.io/gh/iocalebs/patrolbot)

PatrolBot is a Discord bot that generates reports regarding a MediaWiki site's [patrol logs](https://en.wikipedia.org/wiki/Wikipedia:Recent_changes_patrol).

It was initially built for [Zelda Wiki](http://zeldawiki.wiki) but is designed to be usable by any wiki in any language.

![Screenshot of a PatrolBot report posted in a Discord channel](.github/images/report.png)

## Getting Started

1. [Install Go](https://go.dev/doc/install)

2. Install PatrolBot

```sh
go install github.com/iocalebs/patrolbot@latest
```

3. Initialize PatrolBot templates and configuration

```sh
patrolbot init
```

4. Generate a report

```sh
patrolbot report expiring
```

## Configuration

You can customize reports by editing the `config.yaml` file and/or templates in `$HOME/.patrolbot` created by `patrolbot init`.

- See [`Config`](https://pkg.go.dev/github.com/iocalebs/patrolbot/internal/config#Config) type documentation details on supported configuration properties
- See [`ReportData`](https://pkg.go.dev/github.com/iocalebs/patrolbot/internal/report/data#ReportData) type documentation for details on the data available in report templates.
- See [`text/template`](https://pkg.go.dev/text/template) for general information on the Go template syntax used for PatrolBot reports.

A [JSON Schema](https://raw.githubusercontent.com/iocalebs/patrolbot/refs/heads/trunk/config.schema.json) is available for validation and editor autocompletion of `config.yaml`.
