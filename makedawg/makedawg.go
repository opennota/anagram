package main

import (
	"bufio"
	"encoding/binary"
	"os"
	"sort"
	"unicode/utf8"

	"github.com/opennota/dawg"
)

func r2b(r rune) byte {
	if r >= 'а' && r <= 'я' {
		return byte(r - 'а')
	}
	if r == '-' {
		return 32
	}
	if r == '\'' {
		return 33
	}
	panic("unreachable")
}

func conv(s string) string {
	n := utf8.RuneCountInString(s)
	buf := make([]byte, n)
	i := 0
	for _, r := range s {
		buf[i] = r2b(r)
		i++
	}
	return string(buf)
}

func main() {
	f, err := os.Open("words.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	var words []string
	for s.Scan() {
		words = append(words, conv(s.Text()))
	}
	sort.Strings(words)

	var d dawg.DAWG
	for _, w := range words {
		d.Insert(w)
	}
	d.Finish()

	fl := d.Flatten()

	fo, err := os.OpenFile("dawg.bin", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4)
	for _, n := range fl {
		binary.LittleEndian.PutUint32(buf, uint32(n.Index))
		buf[3] = n.B
		if n.EOL {
			buf[3] |= 0x80
		}
		if n.F {
			buf[3] |= 0x40
		}
		fo.Write(buf)
	}
	fo.Close()
}
