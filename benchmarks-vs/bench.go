package benchropes

import (
	"bytes"
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

	bb := b.Bytes()

	// convert to runes to deal with runes

	if at == 0 {
		copy(bb[len(str):], bb[0:])
		copy(bb[0:len(str)], []byte(str))
		return nil
	}

	rs := []rune(string(bb))
	ri := []rune(str)
	size := len(ri)
	if at >= len(rs) {
		rs = append(rs, ri...)
		goto copyback
	}

	// log.Printf("lenrs %d, size: %d. At: %d", len(rs), size, at)
	rs = append(rs, make([]rune, size)...) // +1 coz this is a dummy package and I can't be arsed to actually do correct counts
	// log.Printf("lenrs %d at %d, at+size: %d", len(rs), at, at+size)

	copy(rs[at+size:], rs[at:])
	copy(rs[at:at+size], ri)
copyback:
	bb8 := []byte(string(rs))
	if len(bb8) > len(bb) {
		bb = append(bb, make([]byte, len(bb8)-len(bb))...)
	}
	copy(bb, bb8)
	return nil
}

func (b *naivebuffer) EraseAt(at, n int) error {
	bb := b.Bytes()
	rs := []rune(string(bb))
	copy(rs[at:], rs[at+n:])
	rs = rs[:len(rs)-n]
	bb8 := []byte(string(rs))
	copy(bb, bb8)
	return nil
}
