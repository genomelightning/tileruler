// Tile Ruler is a command line tool for generating PNGs based on given abv files.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Unknwon/com"

	"github.com/genomelightning/tileruler/abv"
	// "github.com/genomelightning/tileruler/rule"
)

var (
	abvPath      = flag.String("abv-path", "./", "directory or path of abv file(s)")
	imgDir       = flag.String("img-dir", "pngs", "path to store PNGs")
	singleMode   = flag.Bool("single-mode", true, "generate PNG per abv")
	slotPixel    = flag.Int("slot-pixel", 1, "slot pixel of width and height")
	hasGrids     = flag.Bool("has-grids", false, "indicates whether slot has border")
	startBandIdx = flag.Int("start-band", 0, "start band index")
	startPosIdx  = flag.Int("start-pos", 0, "start position index")
	maxBandIdx   = flag.Int("max-band", 9, "max band index")
	maxPosIdx    = flag.Int("max-pos", 49, "max position index")
	boxNum       = flag.Int("box-num", 13, "box number")
	workNum      = flag.Int("work-num", 10, "work chan buffer")
)

var start = time.Now()

type Option struct {
	ImgDir      string
	SlotPixel   int
	HasGrids    bool
	IsSingleAbv bool
	*abv.Range
	MaxWorkNum int // Max goroutine number.
}

func validateInput() (*Option, error) {
	flag.Parse()
	opt := &Option{
		ImgDir:      *imgDir,
		SlotPixel:   *slotPixel,
		HasGrids:    *hasGrids,
		IsSingleAbv: *singleMode,
		Range: &abv.Range{
			StartBandIdx: *startBandIdx,
			EndBandIdx:   *maxBandIdx,
			StartPosIdx:  *startPosIdx,
			EndPosIdx:    *maxPosIdx,
		},
		MaxWorkNum: *workNum,
	}

	if opt.HasGrids {
		if opt.SlotPixel < 2 {
			return nil, errors.New("-slot-pixel cannot be smaller than 2 with grids")
		}
	} else if opt.SlotPixel < 1 {
		return nil, errors.New("-slot-pixel cannot be smaller than 1")
	}

	switch {
	case *boxNum < 13:
		log.Fatalln("-box-num cannot be smaller than 13")
	case opt.MaxWorkNum < 1:
		log.Fatalln("-work-num cannot be smaller than 1")
	}
	return opt, nil
}

// getAbvList returns a list of abv file paths.
// It recognize if given path is a file.
func getAbvList(abvPath string) ([]string, error) {
	if !com.IsExist(abvPath) {
		return nil, errors.New("Given abv path does not exist")
	} else if com.IsFile(abvPath) {
		return []string{abvPath}, nil
	}

	// Given path is a directory.
	dir, err := os.Open(abvPath)
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
			abvs = append(abvs, filepath.Join(abvPath, fi.Name()))
		}
	}
	return abvs, nil
}

func rangeString(idx int) string {
	if idx < 0 {
		return "MAX"
	}
	return com.ToStr(idx)
}

func generateAbvImgs(opt *Option, names []string) error {
	for i, name := range names {
		h, err := abv.Parse(name, opt.Range, nil)
		if err != nil {
			log.Fatalf("Fail to parse abv file(%s): %v", name, err)
		}
		h.Name = filepath.Base(name)

		opt.EndBandIdx = h.MaxBand
		opt.EndPosIdx = h.MaxPos
		if err = GenerateAbvImg(opt, h); err != nil {
			return err
		}
		fmt.Printf("%d[%s]: %s: %d * %d\n", i, time.Since(start), h.Name, h.MaxBand, h.MaxPos)
	}
	return nil
}

func main() {
	opt, err := validateInput()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Option:")
	fmt.Printf("Band Range: %d - %s\n", opt.StartBandIdx, rangeString(opt.EndBandIdx))
	fmt.Printf("Pos Range: %d - %s\n", opt.StartPosIdx, rangeString(opt.EndPosIdx))
	fmt.Println("Has Grids:", opt.HasGrids)
	fmt.Println()

	runtime.GOMAXPROCS(runtime.NumCPU())

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

	// Parse abv file.
	var humans []*abv.Human
	names, err := getAbvList(*abvPath)
	if err != nil {
		log.Fatalf("Fail to get abv list: %v", err)
	} else if len(names) == 0 {
		log.Fatalf("No abv files found: %s", *abvPath)
	}
	// humans = make([]*abv.Human, len(names))

	if opt.IsSingleAbv {
		err = generateAbvImgs(opt, names)
	}
	if err != nil {
		log.Fatalf("Fail to generate PNG files: %v", err)
	}
	fmt.Println("Time spent(total):", time.Since(start))
	return

	for i, name := range names {
		humans[i], err = abv.Parse(name, opt.Range, nil)
		if err != nil {
			log.Fatalf("Fail to parse abv file(%s): %v", name, err)
		}
		humans[i].Name = filepath.Base(name)
		fmt.Printf("%d: %s: %d * %d\n", i, humans[i].Name, humans[i].MaxBand, humans[i].MaxPos)
	}
	fmt.Println("Time spent(parse blocks):", time.Since(start))
	fmt.Println()

	// Get max band and position index.
	realMaxBandIdx := -1
	realMaxPosIdx := -1
	for _, h := range humans {
		if h.MaxBand > realMaxBandIdx {
			realMaxBandIdx = h.MaxBand
		}
		if h.MaxPos > realMaxPosIdx {
			realMaxPosIdx = h.MaxPos
		}
		// fmt.Println("Pos Count:", h.PosCount)
	}

	if opt.EndBandIdx < 0 || opt.EndBandIdx > realMaxBandIdx {
		opt.EndBandIdx = realMaxBandIdx
	}
	if opt.EndPosIdx < 0 || opt.EndPosIdx > realMaxPosIdx {
		opt.EndPosIdx = realMaxPosIdx
	}
	fmt.Println("Max Band Index:", opt.EndBandIdx, "\nMax Pos Index:", opt.EndPosIdx)
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
