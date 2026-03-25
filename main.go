// PatrolBot assists with patrolling MediaWiki sites.
package main

import "github.com/iocalebs/patrolbot/internal/cli"

//go:generate go run internal/tools/genschema/genschema.go

func main() {
	cli.Execute()
}

// TODO: Simplify init (no Discord bot)
