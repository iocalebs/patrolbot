// Gentestdata generates real MediaWiki Action API responses,
// which can then be used to mock the API in tests
package main

import (
	"flag"
	"log"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/spf13/pflag"
)

func main() {
	overwrite := flag.Bool("overwrite", false, "overwrite existing testdata files")
	cfgFile := flag.String("config", "../../config.yaml", "path to patrolbot config file")

	flag.Parse()

	cfg, err := config.Load(*cfgFile, &pflag.FlagSet{})
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	err = mwTestData(cfg, *overwrite)
	if err != nil {
		log.Fatal(err)
	}
}
