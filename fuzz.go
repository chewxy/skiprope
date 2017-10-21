// +build gofuzz

package skiprope

import (
	"os"
	"unicode/utf8"
)

func Fuzz(data []byte) int {
	r := New()
	if err := r.Insert(0, string(data)); err != nil {
		return 0
	}
	rb := New()
	if err := rb.InsertBytes(0, data); err != nil {
		return 0
	}

	if r.String() != rb.String() {
		println("Insert and InsertBytes did not insert equally")
		println(r.String())
		println(rb.String())
		os.Exit(1)
	}

	runeData := []rune{}
	for i, w := 0, 0; i < len(data); i += w {
		runeValue, width := utf8.DecodeRune(data[i:])
		runeData = append(runeData, runeValue)
		w = width
	}

	rr := New()
	if err := rr.InsertRunes(0, runeData); err != nil {
		return 0
	}

	// Insert not at = 0
	if err := rb.InsertBytes(len(data)/3, data); err != nil {
		return 0
	}
	_ = rb.String()

	// Before
	_, _, err := rb.Before(len(data)/2, func(r rune) bool { return true })
	if err != nil {
		return 0
	}
	_, _, err = rb.Before(len(data)/2, func(r rune) bool { return false })
	if err != nil {
		return 0
	}
	_, _, err = rb.Before(len(data)/2, func(r rune) bool { return r == '0' })
	if err != nil {
		return 0
	}

	// EraseAt
	eraser := New()
	if err := eraser.InsertBytes(0, data); err != nil {
		return 0
	}
	if err := eraser.EraseAt(len(data)/2, len(data)/3); err != nil {
		return 0
	}

	// Index
	indexer := New()
	if err := indexer.InsertBytes(0, data); err != nil {
		return 0
	}
	indexer.Index(len(data) / 2)

	// Runes
	runer := New()
	if err := runer.InsertBytes(0, data); err != nil {
		return 0
	}
	runer.Runes()

	// Size
	sizer := New()
	if err := sizer.InsertBytes(0, data); err != nil {
		return 0
	}
	sizer.Size()

	// Substr
	substrer := New()
	if err := substrer.InsertBytes(0, data); err != nil {
		return 0
	}
	substrer.Substr(0, len(data)/3)

	// SubstrRunes
	substrRuner := New()
	if err := substrRuner.InsertBytes(0, data); err != nil {
		return 0
	}
	substrRuner.SubstrRunes(0, len(data)/3)

	// SubstrBytes
	substrByter := New()
	if err := substrByter.InsertBytes(0, data); err != nil {
		return 0
	}
	substrByter.SubstrBytes(0, len(data)/3)

	return 1
}
