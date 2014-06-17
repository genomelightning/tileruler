// Package tileruler is a genome tile rule parser of lightning project.
package tileruler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Unknwon/com"
)

// Rule represents a tile rule.
type Rule struct {
	TileId  string `xorm:"VARCHAR(15) UNIQUE(s)"`
	Factor  int    // Common factor.
	Band    int    `xorm:"INDEX"` // Band index.
	Pos     int    // Position index.
	Variant int    `xorm:"UNIQUE(s)"` // Variant index.
}

func powInt(x int, y int) int {
	num := 1
	for i := 0; i < y; i++ {
		num *= x
	}
	return num
}

// hexStr2int converts hex format string to decimal number.
func hexStr2int(hexStr string) (int, error) {
	num := 0
	length := len(hexStr)
	for i := 0; i < length; i++ {
		char := hexStr[length-i-1]
		factor := -1

		switch {
		case char >= '0' && char <= '9':
			factor = int(char) - '0'
		case char >= 'a' && char <= 'f':
			factor = int(char) - 'a' + 10
		default:
			return -1, fmt.Errorf("invalid hex: %s", char)
		}

		num += factor * powInt(16, i)
	}
	return num, nil
}

// parseTileId parses given tile information and returns
// corresponding band and position index.
func parseTileId(info string) (band int, pos int, err error) {
	infos := strings.Split(info, ".")
	if len(infos) != 4 {
		return -1, -1, fmt.Errorf("invalid format")
	}

	band, err = hexStr2int(infos[0])
	if err != nil {
		return -1, -1, fmt.Errorf("cannot parse band index: %v", err)
	}

	pos, err = hexStr2int(infos[2])
	if err != nil {
		return -1, -1, fmt.Errorf("cannot parse position index: %v", err)
	}
	return band, pos, nil
}

// Parse parses a tile rule file and returns all rules.
func Parse(name string) ([]*Rule, error) {
	rules := make([]*Rule, 0, 100)
	if err := IterateParse(name, func(r *Rule) error {
		rules = append(rules, r)
		return nil
	}); err != nil {
		return nil, err
	}
	return rules, nil
}

type IterateFunc func(*Rule) error

func IterateParse(name string, fn IterateFunc) error {
	if !com.IsFile(name) {
		return fmt.Errorf("file(%s) does not exist or is not a file", name)
	}

	fr, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fr.Close()

	lastTileId := ""
	curVarIndex := 0

	var errRead error
	var line string
	buf := bufio.NewReader(fr)
	for idx := 0; errRead != io.EOF; idx++ {
		line, errRead = buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if errRead != nil {
			if errRead != io.EOF {
				return errRead
			}
		}
		if len(line) == 0 {
			break // Nothing left.
		}

		r := new(Rule)
		infos := strings.Split(line, ",")
		r.Factor, err = com.StrTo(infos[0]).Int()
		if err != nil {
			return fmt.Errorf("%d: cannot parse factor of line[%s]: %v", idx, line, err)
		}

		r.TileId = infos[1][1 : len(infos[1])-1]
		r.Band, r.Pos, err = parseTileId(r.TileId)
		if err != nil {
			return fmt.Errorf("%d: cannot parse ID of line[%s]: %v", idx, line, err)
		}

		curVarIndex++
		if r.TileId == lastTileId {
			r.Variant = curVarIndex
		} else {
			curVarIndex = 0
			lastTileId = r.TileId
		}

		if err = fn(r); err != nil {
			return err
		}
	}
	return nil
}
