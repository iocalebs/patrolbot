// Gentestdata generates real MediaWiki Action API responses,
// which can then be used to mock the API in tests
package main

import (
	"log"
)

const testWiki = "zw_en"

func main() {
	err := mwTestData()
	if err != nil {
		log.Fatal(err)
	}
}
