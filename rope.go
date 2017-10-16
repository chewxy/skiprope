// package skiprope provides rope-like data structure for efficient manipulation of large strings.
package skiprope

import (
	"fmt"
)

const (
	MaxHeight  = 60 // maximum size of the skiplist
	BucketSize = 64 // data bucket size in a knot - about 64 runes is optimal for insertion on a core i7.
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
	Head knot
	size int
}

// knot is a node in a rope.... because... geddit?
type knot struct {
	data   [BucketSize]rune // array is preallocated - 64*4 bytes used
	nexts  []skipknot       // next
	height int              // number of elements located in nexts. Minium height is 1
	used   int              // indicates how many runes are used in data
}

func newKnot(height int) *knot {
	return &knot{
		height: height,
		nexts:  make([]skipknot, height),
	}
}

func (k knot) GoString() string {
	return fmt.Sprintf("Data: %q | (Height %d, Used %d) | %v", string(k.data[:]), k.height, k.used, k.nexts)
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

// SubstrRunes is like Substr, but returns []rune
func (r *Rope) SubstrRunes(pointA, pointB int) []rune {
	lastPos := r.size
	a := clamp(min(pointA, pointB), 0, lastPos)
	b := clamp(max(pointB, pointB), 0, lastPos)

	size := b - a
	if size == 0 {
		return nil
	}
	retVal := make([]rune, size)

	s := skiplist{r: r}
	var k1, k2 *knot
	var start, end, retOffset int
	var err error
	if k1, start, _, err = s.find(a); err != nil {
		panic(err)
	}
	if k2, end, _, err = s.find(b); err != nil {
		panic(err)
	}

	ds := start // data start
	for n := k1; n != nil; n = n.nexts[0].knot {
		de := n.used // data end
		// last block
		if n == k2 {
			de = end
		}
		copy(retVal[retOffset:], n.data[ds:de])
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
	return string(r.SubstrRunes(pointA, pointB))
}

// InsertRunes inserts the runes at the point
func (r *Rope) InsertRunes(point int, data []rune) (err error) {
	if point > r.size {
		point = r.size
	}

	// search for the Knot where we'll insert
	var k *knot
	s := skiplist{r: r}
	if k, _, _, err = s.find(point); err != nil {
		return err
	}
	return s.insert(k, data)
}

// Insert inserts the string at the point
func (r *Rope) Insert(at int, str string) error {
	return r.InsertRunes(at, []rune(str))
}

func (r *Rope) EraseAt(point, n int) (err error) {
	if point > r.size {
		point = r.size
	}
	if n >= r.size-point {
		n = r.size - point
	}
	var k *knot
	s := skiplist{r: r}
	if k, _, _, err = s.find(point); err != nil {
		return err
	}
	s.del(k, n)
	return nil
}

func (r *Rope) Index(at int) rune {
	s := skiplist{r: r}
	var k *knot
	var offset int
	var err error

	if k, offset, _, err = s.find(at); err != nil {
		return -1
	}
	if offset == BucketSize {
		return k.nexts[0].data[0]
	}
	return k.data[offset]
}

func (r *Rope) String() string {
	return r.Substr(0, r.size)
}

// Before returns the index of the first rune that matches the function before the given point.
//
// Example: "Hello World". Let's say `at` is at 9 (rune = r). And we want to find the whitespace before it.
// This function will return 5, which is the index of the rune immediately after the whitespace.
func (r *Rope) Before(at int, fn func(r rune) bool) (retVal int, retRune rune, err error) {
	s := skiplist{r: r}
	var k, prev *knot
	var offset int

	if k, offset, _, err = s.find(at); err != nil {
		return -1, -1, err
	}

	if fn(k.data[offset]) {
		return at, k.data[offset], nil
	}

	// if it's within this block, then return immediately, otherwise, get previous block
	var befores int
	for i := len(k.data[:offset]); i >= 0; i-- {
		if fn(k.data[i]) {
			return at - befores, k.data[i], nil
		}
		befores++
	}

	// otherwise we'd have to iterate thru
	befores = offset + 1
	for {
		for it := &r.Head; it != nil; it = it.nexts[0].knot {
			if it == k {
				if prev == nil {
					prev = it
				}
				break
			}
			prev = it
		}

		for i := len(prev.data) - 1; i >= 0; i-- {
			if fn(prev.data[i]) {
				return at - befores, prev.data[i], nil
			}
			befores++
		}
		k = prev
	}
	return
}
