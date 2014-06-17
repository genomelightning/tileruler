// Tile Ruler is a command line tool for parsing genome tile rules.
package main

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Unknwon/com"

	"github.com/genomelightning/tileruler/abv"
	"github.com/genomelightning/tileruler/rule"
)

func initImage() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, 209, 209))
	draw.Draw(m, m.Bounds(), image.White, image.ZP, draw.Src)

	// Draw borders.
	for i := m.Bounds().Min.X; i < m.Bounds().Max.X; i++ {
		m.Set(i, m.Bounds().Min.Y, image.Black)
		m.Set(i, m.Bounds().Max.Y-1, image.Black)
	}
	for i := m.Bounds().Min.Y; i < m.Bounds().Max.Y; i++ {
		m.Set(m.Bounds().Min.X, i, image.Black)
		m.Set(m.Bounds().Max.X-1, i, image.Black)
	}

	// Draw grids.
	for i := 1; i < 13; i++ {
		for j := m.Bounds().Min.Y; j < m.Bounds().Max.Y; j++ {
			m.Set(i*16, j, image.Black)
		}
	}
	for i := 1; i < 13; i++ {
		for j := m.Bounds().Min.X; j < m.Bounds().Max.X; j++ {
			m.Set(j, i*16, image.Black)
		}
	}
	return m
}

func drawSquare(m *image.RGBA, c color.Color, x, y int) {
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			m.Set(x*16+i+1, y*16+j+1, c)
		}
	}
}

func getAbvList() ([]string, error) {
	dir, err := os.Open("/Users/jiahuachen/Downloads/abram")
	if err != nil {
		return nil, err
	}

	fis, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	abvs := make([]string, 0, len(fis))
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), ".abv") {
			abvs = append(abvs, fi.Name())
		}
	}
	return abvs, nil
}

func main() {
	start := time.Now()

	var rules map[int]map[int]map[int]*rule.Rule
	var err error
	if !com.IsExist("tilerules.gob") {
		// Parse tile rules.
		rules, err = rule.Parse("/Users/jiahuachen/Downloads/abram/tiles_w_variants.count.sorted")
		if err != nil {
			log.Fatalf("Fail to parse rule file: %v", err)
		}
		fmt.Println("Time spent(parse rules):", time.Since(start))

		fw, err := os.Create("tilerules.gob")
		if err != nil {
			log.Fatalf("Fail to create gob file: %v", err)
		}
		defer fw.Close()

		if err = gob.NewEncoder(fw).Encode(rules); err != nil {
			log.Fatalf("Fail to encode gob file: %v", err)
		}
		fmt.Println("Time spent(encode gob):", time.Since(start))
	} else {
		fr, err := os.Open("tilerules.gob")
		if err != nil {
			log.Fatalf("Fail to create gob file: %v", err)
		}
		defer fr.Close()

		if err = gob.NewDecoder(fr).Decode(&rules); err != nil {
			log.Fatalf("Fail to decode gob file: %v", err)
		}
		fmt.Println("Time spent(decode gob):", time.Since(start))
	}

	images := make([][]*image.RGBA, 6)
	for i := range images {
		images[i] = make([]*image.RGBA, 101)
		for j := range images[i] {
			images[i][j] = initImage()
		}
	}

	names, err := getAbvList()
	if err != nil {
		log.Fatalf("Fail to get abv list: %v", err)
	}
	humans := make([][]*abv.Block, len(names))

	for i, name := range names {
		humans[i], err = abv.Parse(fmt.Sprintf("/Users/jiahuachen/Downloads/abram/%s", name), rules)
		if err != nil {
			log.Fatalf("Fail to parse abv file(%s): %v", "hu011C57.abv", err)
		}

		for _, b := range humans[i] {
			drawSquare(images[b.Band][b.Pos], color.RGBA{b.Color, b.Color, b.Color, 255}, i%13, i/13)
		}
	}
	fmt.Println("Time spent(parse blocks):", time.Since(start))

	for i := range images {
		for j := range images[i] {
			fr, err := os.Create(fmt.Sprintf("imgs/%d-%d.png", i, j))
			if err != nil {
				log.Fatalf("Fail to create png file: %v", err)
			}
			png.Encode(fr, images[i][j])
			fr.Close()
		}
	}

	fmt.Println("Time spent(total):", time.Since(start))
}
