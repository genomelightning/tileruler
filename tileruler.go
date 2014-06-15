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
	Factor  int // Common factor.
	Band    int `xorm:"INDEX"` // Band index.
	Pos     int // Position index.
	Variant int // Variant index.
}

// parseTileId parses given tile information and returns
// corresponding band and position index.
// It handles quotes around information.
// e.g.: "000.00.001.000"
func parseTileId(info string) (band int, pos int, err error) {
	infos := strings.Split(info[1:len(info)-1], ".")
	if len(infos) != 4 {
		return -1, -1, fmt.Errorf("invalid format")
	}

	band, err = com.StrTo(infos[0]).Int()
	if err != nil {
		return -1, -1, fmt.Errorf("cannot parse band index: %v", err)
	}

	// TODO: convert hex format string to decimal number.
	pos, err = com.StrTo(infos[2]).Int()
	if err != nil {
		return -1, -1, fmt.Errorf("cannot parse position index: %v", err)
	}
	return band, pos, nil
}

// Parse parses a tile rule file and returns all rules.
func Parse(name string) ([]*Rule, error) {
	if !com.IsFile(name) {
		return nil, fmt.Errorf("file(%s) does not exist or is not a file", name)
	}

	fr, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	lastTileId := ""
	curVarIndex := 0

	rules := make([]*Rule, 0, 100)
	var errRead error
	var line string
	buf := bufio.NewReader(fr)
	for idx := 0; errRead != io.EOF; idx++ {
		line, errRead = buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if errRead != nil {
			if errRead != io.EOF {
				return nil, errRead
			}
		} else if len(line) == 0 {
			break // Nothing left.
		}

		r := new(Rule)
		infos := strings.Split(line, ",")
		r.Factor, err = com.StrTo(infos[0]).Int()
		if err != nil {
			return nil, fmt.Errorf("cannot parse factor of line[%s]: %v", line, err)
		}

		tileId := infos[1]
		r.Band, r.Pos, err = parseTileId(tileId)
		if err != nil {
			return nil, fmt.Errorf("cannot parse ID of line[%s]: %v", line, err)
		}

		curVarIndex++
		if tileId == lastTileId {
			r.Variant = curVarIndex
		} else {
			curVarIndex = 0
			lastTileId = tileId
		}

		rules = append(rules, r)
		if idx == 99 {
			break
		}
	}
	return rules, nil
}
