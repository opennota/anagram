// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

//go:generate esc -pkg main -o dawg.go dawg.bin

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	eol      = 0x80
	final    = 0x40
	byteMask = ^byte(eol | final)
	eow      = 1 << 28
	dash     = 1 << 29
	apos     = 1 << 30
	runeMask = ^rune(eow | dash | apos)
)

var dawg []byte

func byteAt(index int) byte {
	return dawg[index+3] & byteMask
}

func isEOL(index int) bool {
	return dawg[index+3]&eol != 0
}

func isFinal(index int) bool {
	return dawg[index+3]&final != 0
}

func next(index int) int {
	return 4 * (int(dawg[index]) +
		int(dawg[index+1])<<8 +
		int(dawg[index+2])<<16)
}

func str(rr []rune) string {
	s := ""
	d := ""
	for _, r := range rr {
		if d != "" {
			s += d
		}
		s += string(r & runeMask)
		if r&eow != 0 {
			d = " "
		} else if r&dash != 0 {
			d = "-"
		} else if r&apos != 0 {
			d = "'"
		} else {
			d = ""
		}
	}
	return s
}

func numWords(rr []rune) int {
	n := 0
	for _, r := range rr {
		if r&eow != 0 {
			n++
		}
	}
	return n
}

func printAnagramsInternal(w io.Writer, index, level, maxWords int, runeset, tmp []rune, remaining []int, orig string) {
	if level == len(tmp) {
		anagram := str(tmp)
		if anagram != orig {
			fmt.Fprintln(w, anagram)
		}
		return
	}

	if numWords(tmp[:level]) >= maxWords {
		return
	}

	for {
		b := byteAt(index)

		if b == 32 || b == 33 {
			var mark rune = dash
			if b == 33 {
				mark = apos
			}
			tmp[level-1] |= mark
			printAnagramsInternal(w, next(index), level, maxWords, runeset, tmp, remaining, orig)
			tmp[level-1] &= ^mark
		}

		for i, r := range runeset {
			if remaining[i] == 0 || b != byte(r-'а') {
				continue
			}

			remaining[i]--
			tmp[level] = r
			if nextIndex := next(index); nextIndex != 0 && level+1 < len(tmp) {
				printAnagramsInternal(w, nextIndex, level+1, maxWords, runeset, tmp, remaining, orig)
			}
			if isFinal(index) {
				tmp[level] |= eow
				printAnagramsInternal(w, 0, level+1, maxWords, runeset, tmp, remaining, orig)
			}
			remaining[i]++
		}

		if isEOL(index) {
			break
		}

		index += 4
	}
}

func printAnagrams(w io.Writer, s string, maxWords int) {
	s = strings.TrimSpace(strings.ToLower(s))
	m := map[rune]int{}
	n := 0
	for _, r := range s {
		if r == 'ё' {
			r = 'е'
		} else if r < 'а' || r > 'я' {
			continue
		}
		m[r]++
		n++
	}
	runeset := make([]rune, 0, len(m))
	remaining := make([]int, 0, len(m))
	for k, v := range m {
		runeset = append(runeset, k)
		remaining = append(remaining, v)
	}
	tmp := make([]rune, n)
	printAnagramsInternal(w, 0, 0, maxWords, runeset, tmp, remaining, s)
}

func init() {
	dawg, _ = FSByte(false, "/dawg.bin")
}

func main() {
	n := 1
	if len(os.Args) > 2 {
		m, err := strconv.Atoi(os.Args[2])
		if err == nil {
			n = m
		}
	}
	printAnagrams(os.Stdout, os.Args[1], n)
}
