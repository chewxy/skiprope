// package skiprope provides rope-like data structure for efficient manipulation of large strings.
package skiprope

import (
	"fmt"
	"unicode/utf8"
)

const (
	MaxHeight  = 60 // maximum size of the skiplist
	BucketSize = 64 // data bucket size in a knot - about 64 bytes is optimal for insertion on a core i7.
)

// Bias indicates the probability that a new knot will have height of n+1.
// This is the parameter to tweak when considering the tradeoff between high amounts of append operations
// and amount of random writes.
//
// The higher the bias is, the better the data structure is at performing end-of-string appends. The tradeoff is
// performance of random writes will deterioriate.
var Bias = 20

// Rope is a rope data structure built on top of a skip list.
type Rope struct {
	Head  knot
	size  int // number of bytes
	runes int // number of code points
}

// knot is a node in a rope.... because... geddit?
type knot struct {
	data   [BucketSize]byte // array is preallocated - 64*4 bytes used
	nexts  []skipknot       // next
	height int              // number of elements located in nexts. Minium height is 1
	used   int              // indicates how many byte are used in data
}

func newKnot(height int) *knot {
	return &knot{
		height: height,
		nexts:  make([]skipknot, height),
	}
}

type skipknot struct {
	*knot
	skipped      int // number of bytes between the start and current node and the start of the next
	skippedRunes int // number of runes between the start and current node and the start of the next
}

func (s skipknot) GoString() string {
	return fmt.Sprintf("%#v | skipped %v", s.knot, s.skipped)
}

// New creates a new Rope.
func New() *Rope {
	r := &Rope{
		Head: knot{
			height: 1,
			nexts:  make([]skipknot, MaxHeight),
		},
	}
	return r
}

// Size is the length of the rope.
func (r *Rope) Size() int { return r.size }

// Runes is the number of runes in the rope
func (r *Rope) Runes() int { return r.runes }

// SubstrRunes is like Substr, but returns []rune
func (r *Rope) SubstrRunes(pointA, pointB int) []rune {
	return []rune(string(r.SubstrBytes(pointA, pointB)))
}

// SubstrBytes returns a byte slice given the "substring" of the rope.
// Both pointA and pointB refers to the rune, not the byte offset.
//
// Example: "你好world" has a length of 11 bytes. If we only want "好",
// we'd have to call SubstrBytes(1, 2), not SubstrBytes(3, 6) which you would if you
// were dealing with pure bytes
func (r *Rope) SubstrBytes(pointA, pointB int) []byte {
	lastPos := r.runes
	a := clamp(min(pointA, pointB), 0, lastPos)
	b := clamp(max(pointB, pointB), 0, lastPos)

	size := b - a
	if size == 0 {
		return nil
	}
	s := skiplist{r: r}
	var k1, k2 *knot
	var start, end, retOffset, startSkipped, endSkipped int
	var err error
	if k1, start, startSkipped, err = s.find(a); err != nil {
		panic(err)
	}
	if k2, end, endSkipped, err = s.find(b); err != nil {
		panic(err)
	}

	retVal := make([]byte, 0, (endSkipped+end)-(start+startSkipped))
	ds := start // data start
	for n := k1; n != nil; n = n.nexts[0].knot {
		de := n.used // data end
		// last block
		if n == k2 {
			de = end
		}

		// copy(retVal[retOffset:], n.data[ds:de])
		retVal = append(retVal, n.data[ds:de]...)
		retOffset += (de - ds)
		if n == k1 {
			ds = 0
		}
		if n == k2 {
			break
		}
	}
	return retVal
}

// Substr creates a substring.
func (r *Rope) Substr(pointA, pointB int) string {
	return string(r.SubstrBytes(pointA, pointB))
}

// InsertRunes inserts the runes at the point
func (r *Rope) InsertRunes(point int, data []rune) (err error) {
	return r.InsertBytes(point, []byte(string(data)))
}

// InsertBytes inserts the bytes at the point.
func (r *Rope) InsertBytes(point int, data []byte) (err error) {
	if point > r.runes {
		point = r.runes
	}

	// search for the Knot where we'll insert
	var k *knot
	s := skiplist{r: r}
	if k, err = s.find2(point); err != nil {
		return err
	}
	return s.insert(k, data)
}

// Insert inserts the string at the point
func (r *Rope) Insert(at int, str string) error {
	return r.InsertBytes(at, []byte(str))
}

// EraseAt erases n runes starting from the point.
func (r *Rope) EraseAt(point, n int) (err error) {
	if point > r.runes {
		point = r.runes
	}
	if n >= r.runes-point {
		n = r.runes - point
	}
	var k *knot
	s := skiplist{r: r}
	if k, err = s.find2(point); err != nil {
		return err
	}
	s.del(k, n)
	return nil
}

// Index returns the rune at the given index.
func (r *Rope) Index(at int) rune {
	s := skiplist{r: r}
	var k *knot
	var offset int
	var err error

	if k, offset, _, err = s.find(at); err != nil {
		return -1
	}
	if offset == BucketSize {
		char, _ := utf8.DecodeRune(k.nexts[0].data[0:])
		return char
	}
	char, _ := utf8.DecodeRune(k.data[offset:])
	return char
}

// String returns the rope as a full string.
func (r *Rope) String() string {
	return r.Substr(0, r.runes)
}

// Before returns the index of the first rune that matches the function before the given point.
//
// Example: "Hello World". Let's say `at` is at 9 (rune = r). And we want to find the whitespace before it.
// This function will return 5, which is the byte index of the rune immediately after the whitespace.
func (r *Rope) Before(at int, fn func(r rune) bool) (retVal int, retRune rune, err error) {
	s := skiplist{r: r}
	var k, prev *knot
	var offset int
	var char rune

	if k, offset, _, err = s.find(at); err != nil {
		return -1, -1, err
	}

	char, _ = utf8.DecodeRune(k.data[offset:])
	if fn(char) {
		return at, char, nil
	}

	// if it's within this block, then return immediately, otherwise, get previous block
	var befores, start int
	size := 1
	for end := len(k.data[:offset]); end > start; end -= size {
		char, size = utf8.DecodeLastRune(k.data[start:end])
		if fn(char) {
			return at - befores, char, nil
		}
		befores += size
	}

	// otherwise we'd have to iterate thru the blocks
	befores = offset + 1
	for {
		for it := &r.Head; it != nil; it = it.nexts[0].knot {
			if it == k {
				// if prev == nil {
				// 	prev = it
				// }
				break
			}
			prev = it
		}
		start = 0
		size = 1
		for end := len(prev.data) - 1; end > start; end -= size {
			char, size = utf8.DecodeLastRune(prev.data[start:end])
			if fn(char) {
				return at - befores, char, nil
			}
			befores += size
		}
		k = prev
		// if k == &r.Head {
		// 	break
		// }
	}
	return
}
