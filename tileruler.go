// Tile Ruler is a command line tool for parsing genome tile rules.
package main

import (
	// "encoding/gob"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	// "github.com/Unknwon/com"

	"github.com/genomelightning/tileruler/abv"
	// "github.com/genomelightning/tileruler/rule"
)

var (
	abvDir       = flag.String("abv-dir", "./", "directory that contains abv files")
	blocksFile   = flag.String("blocks-file", "blocks.gob", "path of blocks gob file")
	imgDir       = flag.String("img-dir", "pngs", "path to store PNGs")
	startBandIdx = flag.Int("start-band", 0, "start band index")
	startPosIdx  = flag.Int("start-pos", 0, "start position index")
	maxBandIdx   = flag.Int("max-band", 9, "max band index")
	maxPosIdx    = flag.Int("max-pos", 49, "max position index")
	slotPixel    = flag.Int("slot-pixel", 16, "slot pixel of width and height")
	boxNum       = flag.Int("box-num", 13, "box number")
	workNum      = flag.Int("work-num", 10, "work chan buffer")
)

func initImage() *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, *boxNum**slotPixel+1, *boxNum**slotPixel+1))
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
	for i := 1; i < *boxNum; i++ {
		for j := m.Bounds().Min.Y; j < m.Bounds().Max.Y; j++ {
			m.Set(i**slotPixel, j, image.Black)
		}
	}
	for i := 1; i < *boxNum; i++ {
		for j := m.Bounds().Min.X; j < m.Bounds().Max.X; j++ {
			m.Set(j, i**slotPixel, image.Black)
		}
	}
	return m
}

var varColors = []color.Color{
	color.RGBA{0, 153, 0, 255},
	color.RGBA{0, 204, 0, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 255, 255, 255},
	color.RGBA{0, 204, 255, 255},
	color.RGBA{0, 153, 255, 255},
	color.RGBA{0, 102, 255, 255},
	color.RGBA{0, 51, 255, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{0, 0, 102, 255},
}

func drawSquare(m *image.RGBA, idx, x, y int) {
	if idx >= len(varColors) {
		idx = len(varColors) - 1
	}

	for i := 0; i < *slotPixel-1; i++ {
		for j := 0; j < *slotPixel-1; j++ {
			m.Set(x**slotPixel+i+1, y**slotPixel+j+1, varColors[idx])
		}
	}
}

func getAbvList(dirPath string) ([]string, error) {
	dir, err := os.Open(dirPath)
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
	flag.Parse()

	// Validate input.
	switch {
	case *slotPixel < 2:
		log.Fatalln("-slot-pixel cannot be smaller than 2")
	case *boxNum < 13:
		log.Fatalln("-box-num cannot be smaller than 13")
	case *workNum < 1:
		log.Fatalln("-work-num cannot be smaller than 1")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	// NOTE: not doing it for now, maybe next stage of server.
	// var rules map[int]map[int]map[int]*rule.Rule
	// var err error
	// if !com.IsExist("tilerules.gob") {
	// 	// Parse tile rules.
	// 	rules, err = rule.Parse("/Users/jiahuachen/Downloads/abram/tiles_w_variants.count.sorted")
	// 	if err != nil {
	// 		log.Fatalf("Fail to parse rule file: %v", err)
	// 	}
	// 	fmt.Println("Time spent(parse rules):", time.Since(start))

	// 	fw, err := os.Create("tilerules.gob")
	// 	if err != nil {
	// 		log.Fatalf("Fail to create gob file: %v", err)
	// 	}
	// 	defer fw.Close()

	// 	if err = gob.NewEncoder(fw).Encode(rules); err != nil {
	// 		log.Fatalf("Fail to encode gob file: %v", err)
	// 	}
	// 	fmt.Println("Time spent(encode gob):", time.Since(start))
	// } else {
	// 	fr, err := os.Open("tilerules.gob")
	// 	if err != nil {
	// 		log.Fatalf("Fail to create gob file: %v", err)
	// 	}
	// 	defer fr.Close()

	// 	if err = gob.NewDecoder(fr).Decode(&rules); err != nil {
	// 		log.Fatalf("Fail to decode gob file: %v", err)
	// 	}
	// 	fmt.Println("Time spent(decode gob):", time.Since(start))
	// }

	var humans []*abv.Human
	// if !com.IsExist(*blocksFile) {
	names, err := getAbvList(*abvDir)
	if err != nil {
		log.Fatalf("Fail to get abv list: %v", err)
	} else if len(names) == 0 {
		log.Fatalf("No abv files found: %s", *abvDir)
	}
	humans = make([]*abv.Human, len(names))

	for i, name := range names {
		humans[i], err = abv.Parse(path.Join(*abvDir, name),
			&abv.Range{*startBandIdx, *maxBandIdx, *startPosIdx, *maxPosIdx}, nil)
		if err != nil {
			log.Fatalf("Fail to parse abv file(%s): %v", name, err)
		}
		// fmt.Printf("%s: %d * %d\n", name, humans[i].MaxBand, humans[i].MaxPos)
	}
	fmt.Println("Time spent(parse blocks):", time.Since(start))

	// fw, err := os.Create(*blocksFile)
	// if err != nil {
	// 	log.Fatalf("Fail to create blocks gob file: %v", err)
	// }
	// defer fw.Close()

	// if err = gob.NewEncoder(fw).Encode(humans); err != nil {
	// 	log.Fatalf("Fail to encode blocks gob file: %v", err)
	// }
	// fmt.Println("Time spent(encode blocks gob):", time.Since(start))
	// } else {
	// 	fr, err := os.Open(*blocksFile)
	// 	if err != nil {
	// 		log.Fatalf("Fail to open blocks gob file: %v", err)
	// 	}
	// 	defer fr.Close()

	// 	if err = gob.NewDecoder(fr).Decode(&humans); err != nil {
	// 		log.Fatalf("Fail to decode blocks gob file: %v", err)
	// 	}
	// 	fmt.Println("Time spent(decode blocks gob):", time.Since(start))
	// }

	realMaxBandIdx := -1
	realMaxPosIdx := -1
	// Get max band and position index.
	for _, h := range humans {
		if h.MaxBand > realMaxBandIdx {
			realMaxBandIdx = h.MaxBand
		}
		if h.MaxPos > realMaxPosIdx {
			realMaxPosIdx = h.MaxPos
		}
		// fmt.Println("Pos Count:", h.PosCount)
	}
	fmt.Println("Max Band Index:", realMaxBandIdx, "Max Pos Index:", realMaxPosIdx)

	if *maxBandIdx < 0 || *maxBandIdx > realMaxBandIdx {
		*maxBandIdx = realMaxBandIdx
	}

	wg := &sync.WaitGroup{}
	workChan := make(chan bool, *workNum)

	os.MkdirAll(*imgDir, os.ModePerm)
	for i := *startBandIdx; i <= *maxBandIdx; i++ {
		// fmt.Println(i)
		wg.Add(*maxPosIdx - *startPosIdx + 1)
		os.MkdirAll(fmt.Sprintf("%s/%d", *imgDir, i), os.ModePerm)
		for j := *startPosIdx; j <= *maxPosIdx; j++ {
			m := initImage()
			for k := range humans {
				if b, ok := humans[k].Blocks[i][j]; ok {
					drawSquare(m, int(b.Variant), k%*boxNum, k / *boxNum)
				}
			}
			workChan <- true
			go func(band, pos int) {
				fr, err := os.Create(fmt.Sprintf("%s/%d/%d.png", *imgDir, band, pos))
				// fr, err := os.Create(fmt.Sprintf("%s/%d/%d.png", *imgDir, i, j))
				if err != nil {
					log.Fatalf("Fail to create png file: %v", err)
				} else if err = png.Encode(fr, m); err != nil {
					log.Fatalf("Fail to encode png file: %v", err)
				}
				fr.Close()
				wg.Done()
				<-workChan
			}(i, j)
		}
		runtime.GC()
	}

	fmt.Println("Goroutine #:", runtime.NumGoroutine())
	wg.Wait()
	fmt.Println("Time spent(total):", time.Since(start))
	return
	// images := make([][]*image.RGBA, realMaxBandIdx+1)
	// for i := range images {
	// 	images[i] = make([]*image.RGBA, realMaxPosIdx+1)
	// 	for j := range images[i] {
	// 		images[i][j] = initImage()
	// 	}
	// }

	// for i := range humans {
	// 	for _, b := range humans[i].Blocks {
	// 		drawSquare(images[b.Band][b.Pos], b.Variant, i%13, i/13)
	// 	}
	// }
	// fmt.Println("Time spent(draw blocks):", time.Since(start))

	// for i := range images {
	// 	for j := range images[i] {
	// 		fr, err := os.Create(fmt.Sprintf("%s/%d-%d.png", *imgDir, i, j))
	// 		if err != nil {
	// 			log.Fatalf("Fail to create png file: %v", err)
	// 		} else if err = png.Encode(fr, images[i][j]); err != nil {
	// 			log.Fatalf("Fail to encode png file: %v", err)
	// 		}
	// 		fr.Close()
	// 	}
	// }

	fmt.Println("Time spent(total):", time.Since(start))
}
