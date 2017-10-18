package benchropes

import (
	"bytes"
	"unicode/utf8"

	er "github.com/eugene-eeo/rope"
)

type eugenerope struct {
	er.Rope
}

func (e *eugenerope) Insert(at int, str string) error {
	if at == 0 {
		pre := er.L(str)
		e.Rope = pre.Concat(e.Rope)
		return nil
	}
	if at >= e.Length() {
		post := er.L(str)
		e.Rope = e.Rope.Concat(post)
		return nil
	}
	i, j := e.SplitAt(at)
	insert := er.L(str)
	e.Rope = er.Concat(i, insert, j)
	return nil
}

func (e *eugenerope) EraseAt(at, n int) error {
	i, j := e.SplitAt(at)
	_, j = j.SplitAt(n)
	e.Rope = i.Concat(j)
	return nil
}

type naivebuffer struct {
	*bytes.Buffer
}

func (b *naivebuffer) Insert(at int, str string) error {
	// ensure we alwys have space
	if b.Cap() < b.Len()+len(str) {
		b.Grow(len(str))
	}
	bs := []byte(str)
	b.Write(bs) // just to add to the length

	bb := b.Bytes()
	offset := byteOffset(bb, at)
	copy(bb[offset+len(str):], bb[offset:])
	copy(bb[offset:offset+len(str)], bs)
	return nil
}

func (b *naivebuffer) EraseAt(at, n int) error {
	bb := b.Bytes()
	offset := byteOffset(bb, at)
	copy(bb[offset:], bb[offset+n:])
	b.Truncate(len(bb) - n)
	return nil
}

// byteOffset takes a slice of bytes, and returns the index at which the expected number of runes there is
func byteOffset(a []byte, runes int) (offset int) {
	if runes == 0 {
		return 0
	}

	var runeCount int
	for _, size := utf8.DecodeRune(a[offset:]); offset < len(a) && runeCount < runes; _, size = utf8.DecodeRune(a[offset:]) {
		offset += size
		runeCount++
	}
	return offset
}
