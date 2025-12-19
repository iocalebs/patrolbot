// Gentestdata generates real MediaWiki Action API responses,
// which can then be used to mock the API in tests
package main

import (
	"flag"
	"log"
)

const testWiki = "zw_en"

func main() {
	overwrite := flag.Bool("overwrite", false, "overwrite existing testdata files")

	flag.Parse()

	err := mwTestData(*overwrite)
	if err != nil {
		log.Fatal(err)
	}
}
