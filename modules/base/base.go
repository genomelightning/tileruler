package base

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"strings"

	"github.com/curoverse/lightning/experimental/tileruler/modules/cli"
	"github.com/curoverse/lightning/experimental/tileruler/modules/log"
)

type Mode int

const (
	SINGLE Mode = iota + 1
	FULL_SIZE
	TRANSPARENT
)

type Range struct {
	EndBandIdx int
	EndPosIdx  int
}

type Option struct {
	ImgDir    string
	Mode      Mode
	AbvPath   string
	ColorSpec string
	*Range
	MaxColIdx   int
	BoxNum      int
	SlotPixel   int
	Border      int
	Force       bool
	CountOnly   bool
	ReversePath string
	WindowSize  int
	HttpPort    string
	FastjPath   string
	RefLibPath  string
}

// ParseOption parses command arguments into Option sutrct.
func ParseOption(ctx *cli.Context) Option {
	opt := Option{
		Mode:      Mode(ctx.Int("mode")),
		ImgDir:    ctx.String("img-dir"),
		AbvPath:   ctx.String("abv-path"),
		ColorSpec: ctx.String("color-spec"),
		Range: &Range{
			EndBandIdx: ctx.Int("max-band"),
			EndPosIdx:  ctx.Int("max-pos"),
		},
		MaxColIdx:   ctx.Int("max-col"),
		BoxNum:      ctx.Int("box-num"),
		SlotPixel:   ctx.Int("slot-pixel"),
		Border:      ctx.Int("border"),
		Force:       ctx.Bool("force"),
		CountOnly:   ctx.Bool("count-only"),
		ReversePath: ctx.String("reverse-path"),
		WindowSize:  ctx.Int("size"),
		HttpPort:    ctx.String("http-port"),
		FastjPath:   ctx.String("fastj-path"),
		RefLibPath:  ctx.String("lib-path"),
	}

	switch {
	case ctx.Command.Name != "stat" &&
		(opt.Mode == FULL_SIZE || opt.Mode == TRANSPARENT) && opt.BoxNum < 13:
		log.Fatal("-box-num cannot be smaller than 13 in full size or transparent mode")
	}
	return opt
}

var (
	Gray      = image.NewUniform(color.RGBA{230, 230, 230, 255})
	PoundGray = image.NewUniform(color.RGBA{230, 230, 231, 255}) // #
)

// Make large enough to store and being able to convert back to abv file.
const DefaultVarColors = `255, 255, 255
0, 204, 0
0, 255, 0
0, 255, 255
0, 204, 255
0, 153, 255
0, 102, 255
0, 51, 255
0, 0, 255
0, 1, 255
0, 2, 255
0, 3, 255
0, 4, 255
0, 5, 255
0, 6, 255
0, 7, 255
0, 8, 255
0, 9, 255
0, 10, 255
0, 11, 255
0, 12, 255
0, 13, 255
0, 14, 255
0, 15, 255
0, 16, 255
0, 17, 255
0, 18, 255
0, 19, 255
0, 20, 255
0, 21, 255
0, 22, 255
0, 23, 255
0, 24, 255
0, 25, 255
0, 26, 255
0, 27, 255
0, 28, 255
0, 29, 255
0, 30, 255
0, 31, 255
0, 32, 255
0, 33, 255
0, 34, 255
0, 35, 255
0, 36, 255
0, 37, 255
0, 38, 255
0, 39, 255
0, 40, 255
0, 41, 255
0, 42, 255
0, 43, 255
0, 44, 255
0, 45, 255
0, 46, 255
0, 47, 255
0, 48, 255
0, 49, 255
0, 50, 255
0, 51, 255
0, 52, 255
0, 53, 255`

var VarColors = make([]color.Color, 0, 20)

// GetVarColorIdx returns index of given variant color based on current color map.
// It returns -1 when it's the background color, -2 when no match.
func GetVarColorIdx(c color.Color) int {
	r, g, b, a := c.RGBA()
	var vr, vg, vb, va uint32 // Declare once to save cost.

	// Compare to background color first.
	vr, vg, vb, va = Gray.RGBA()
	if r == vr && g == vg && b == vb && a == va {
		return -1
	}

	// Try to match on color map.
	for i, vc := range VarColors {
		vr, vg, vb, va = vc.RGBA()
		if r == vr && g == vg && b == vb && a == va {
			return i
		}
	}

	// See if it's a pound gray.
	vr, vg, vb, va = PoundGray.RGBA()
	if r == vr && g == vg && b == vb && a == va {
		return 99
	}

	return -2
}

func parseVarColors(str string) error {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		infos := strings.Split(line, ",")
		if len(infos) < 3 {
			return fmt.Errorf("Not enough color assigned in line[%d]: %s", i, line)
		}
		VarColors = append(VarColors, color.RGBA{
			StrTo(strings.TrimSpace(infos[0])).MustUint8(),
			StrTo(strings.TrimSpace(infos[1])).MustUint8(),
			StrTo(strings.TrimSpace(infos[2])).MustUint8(), 255})
	}
	return nil
}

// ParseColorSpec parses color map based on given file,
// uses default color map if path is empty.
func ParseColorSpec(specPath string) error {
	if IsFile(specPath) {
		spec, err := ioutil.ReadFile(specPath)
		if err != nil {
			return err
		}
		parseVarColors(string(spec))
	} else {
		parseVarColors(DefaultVarColors)
	}
	return nil
}
