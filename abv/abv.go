// Package abv parses abv format files based on tile rules.
package abv

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/Unknwon/com"

	"github.com/genomelightning/tileruler/rule"
)

// Block represents a block for a human in given position in slippy map.
type Block struct {
	Band  int   // Band index.
	Pos   int   // Position index.
	Color uint8 // NOTE: might change to int for special meaning.
}

var encodeStd = []byte("CDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

// Parse parses a abv file based on given tile rules and returns all blocks.
func Parse(name string, rules map[int]map[int]map[int]*rule.Rule) ([]*Block, error) {
	if !com.IsFile(name) {
		return nil, fmt.Errorf("file(%s) does not exist or is not a file", name)
	}

	fr, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	blocks := make([]*Block, 0, 3000000)
	var bandIdx int // Current band index.

	var line []byte
	buf := bufio.NewReader(fr)

	// To skip header e.g.: "huFE71F3"
	_, errRead := buf.ReadBytes(' ')
	if errRead != nil {
		return nil, errRead
	}

	// True for next thing read will be actual body not band index.
	isInBody := false

	for errRead != io.EOF {
		line, errRead = buf.ReadBytes(' ')
		line = bytes.TrimSpace(line)

		if errRead != nil {
			if errRead != io.EOF {
				return nil, errRead
			}
		}
		if len(line) == 0 {
			break
		}

		if !isInBody {
			bandIdx, err = com.HexStr2int(string(line))
			if err != nil {
				return nil, err
			}

			// NOTE: limit band and position just for debugging purpose.
			if bandIdx > 5 {
				break
			}
		} else {
			for i, char := range line {
				// NOTE: limit band and position just for debugging purpose.
				if i > 100 {
					break
				}

				varIdx := -1
				switch char {
				case '-', '#': // Not recognize or just skip.
					continue
				case '.': // Default variant.
					varIdx = 0
				default:
					// Non-default variant.
					varIdx = bytes.IndexByte(encodeStd, char)
					if varIdx < 1 {
						return nil, fmt.Errorf("Invalid version of variant[%s]: %s", line, string(char))
					}
				}

				b := &Block{bandIdx, i, 0}
				r, ok := rules[b.Band][b.Pos][varIdx]
				if !ok {
					return nil, fmt.Errorf("Rule not found: %d.%d.%d", b.Band, b.Pos, varIdx)
				}
				b.Color = uint8(r.Factor)
				blocks = append(blocks, b)
			}
		}
		isInBody = !isInBody
	}

	return blocks, nil
}
