// Gentestdata generates real MediaWiki Action API responses,
// which can then be used to mock the API in tests
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/pflag"
)

func main() {
	flagset := pflag.NewFlagSet("gentestdata", pflag.ExitOnError)
	overwrite := flagset.Bool("overwrite", false, "overwrite existing testdata files")
	flagset.String("config", "../../../config.yaml", "path to patrolbot config file")

	err := flagset.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing flags: %v\n", err)
	}

	cfg, err := config.Load(flagset)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working dir: %v", err)
	}

	switch filepath.Base(wd) {
	case "mediawiki":
		err = mwTestData(cfg, *overwrite)
		if err != nil {
			log.Fatal(err)
		}
	case "discord":
		err = discordTestData(cfg, *overwrite)
		if err != nil {
			log.Fatal(err)
		}
	}
}
