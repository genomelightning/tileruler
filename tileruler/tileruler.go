// Tile Ruler is a command line tool for parsing genome tile rules into database format.
package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-xorm/xorm"

	"github.com/genomelightning/tileruler"
)

var x *xorm.Engine

func init() {
	var err error
	// x, err = xorm.NewEngine("sqlite3", "/Users/jiahuachen/Downloads/tilerules.db")
	x, err = xorm.NewEngine("mysql", "root:root@tcp(127.0.0.1:3306)/tilerules?charset=utf8")
	if err != nil {
		log.Fatalf("Fail to create ORM engine: %v", err)
	}

	if err = x.Sync(new(tileruler.Rule)); err != nil {
		log.Fatalf("Fail to sync tables: %v", err)
	}
}

func newRule(r *tileruler.Rule) error {
	has, err := x.Get(&tileruler.Rule{
		TileId:  r.TileId,
		Variant: r.Variant,
	})
	if err != nil {
		return err
	} else if has {
		return nil
	}
	_, err = x.Insert(r)
	return err
}

func main() {
	start := time.Now()
	count := 0

	if err := tileruler.IterateParse(
		"/Users/jiahuachen/Downloads/abram/tiles_w_variants.count.sorted",
		func(r *tileruler.Rule) error {
			count++
			if count < 4200000 {
				return nil
			}
			if count%1000 == 0 {
				runtime.GC()
				fmt.Println("Count:", count)
			}
			return newRule(r)
		}); err != nil {
		log.Fatalf("Fail to parse rule file: %v", err)
	}

	fmt.Println("Total rules:", count)
	fmt.Println("Time spent:", time.Since(start))
}
