# PatrolBot

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
patrolbot report weekly
```

## Configuration

You can configure your own reports, or modify the pre-built ones, by editing the configuration file and/or templates in `$HOME/.patrolbot`.

<!-- TODO: config documentation -->
