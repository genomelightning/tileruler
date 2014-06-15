// Tile Ruler is a command line tool for parsing genome tile rules into database format.
package main

import (
	"fmt"
	"log"

	"github.com/genomelightning/tileruler"
)

func main() {
	rules, err := tileruler.Parse("/Users/jiahuachen/Downloads/abram/tiles_w_variants.count.sorted")
	if err != nil {
		log.Fatalf("Fail to parse rule file: %v", err)
	}

	for _, r := range rules {
		fmt.Println(r)
	}
}
