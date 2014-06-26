package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/genomelightning/tileruler/abv"
)

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

func initImage(opt *Option) *image.RGBA {
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

	if opt.HasGrids {
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
	}
	return m
}

func drawSquare(opt *Option, m *image.RGBA, idx, x, y int) {
	// In case variant number is too large.
	if idx >= len(varColors) {
		idx = len(varColors) - 1
	}

	// Left border offset.
	offset := 0
	if opt.HasGrids {
		offset = 1
	}
	for i := 0; i < opt.SlotPixel; i++ {
		for j := 0; j < opt.SlotPixel; j++ {
			m.Set(x*opt.SlotPixel+i+offset, y*opt.SlotPixel+j+offset, varColors[idx])
		}
	}
}

// GenerateImgPerTile generates one PNG for each tile.
func GenerateImgPerTile(opt *Option, humans []*abv.Human) {
	wg := &sync.WaitGroup{}
	workChan := make(chan bool, opt.MaxWorkNum)

	os.MkdirAll(opt.ImgDir, os.ModePerm)
	for i := opt.StartBandIdx; i <= opt.EndBandIdx; i++ {
		// fmt.Println(i)
		wg.Add(opt.EndPosIdx - opt.StartPosIdx + 1)
		os.MkdirAll(fmt.Sprintf("%s/%d", opt.ImgDir, i), os.ModePerm)
		for j := opt.StartPosIdx; j <= opt.EndPosIdx; j++ {
			m := initImage(opt)
			for k := range humans {
				if b, ok := humans[k].Blocks[i][j]; ok {
					drawSquare(opt, m, int(b.Variant), k%*boxNum, k / *boxNum)
				}
			}
			workChan <- true
			go func(band, pos int) {
				if pos%1000 == 0 {
					fmt.Println(band, pos)
				}
				fr, err := os.Create(fmt.Sprintf("%s/%d/%d.png", opt.ImgDir, band, pos))
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
}

func initAbvImage(opt *Option) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, (opt.EndPosIdx+1)*opt.SlotPixel, (opt.EndBandIdx+1)*opt.SlotPixel))
	draw.Draw(m, m.Bounds(), image.White, image.ZP, draw.Src)
	return m
}

// GenerateAbvImg generates one PNG for each abv file.
func GenerateAbvImg(opt *Option, h *abv.Human) error {
	os.MkdirAll(opt.ImgDir, os.ModePerm)

	m := initAbvImage(opt)
	for i := opt.StartBandIdx; i <= opt.EndBandIdx; i++ {
		for j := opt.StartPosIdx; j <= opt.EndPosIdx; j++ {
			if b, ok := h.Blocks[i][j]; ok {
				drawSquare(opt, m, int(b.Variant), j, i)
			}
		}
	}

	fr, err := os.Create(fmt.Sprintf("%s/%s.png", opt.ImgDir, h.Name))
	if err != nil {
		log.Fatalf("Fail to create PNG file(%s): %v", h.Name, err)
	} else if err = png.Encode(fr, m); err != nil {
		log.Fatalf("Fail to encode PNG file(%s): %v", h.Name, err)
	}
	fr.Close()
	runtime.GC()
	return nil
}
