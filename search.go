package skiprope

import (
	"unicode/utf8"

	"errors"
	// "log"
)

// skiplist is a data structure for searching... it's the "skiplist" part of things
type skiplist struct {
	r *Rope
	s [MaxHeight]skipknot
}

// newKnot will accept a []byte of BucketSize or less
func (s *skiplist) newKnot(data []byte, runeCount int) {
	maxHeight := s.r.Head.height
	newHeight := randInt()

	byteCount := len(data)
	k := newKnot(newHeight)
	k.used = byteCount
	copy(k.data[0:], data)

	// the rest of the reason why anyone bothers to take accounting classes
	for maxHeight <= newHeight {
		s.r.Head.height++
		s.r.Head.nexts[maxHeight] = s.r.Head.nexts[maxHeight-1]

		s.s[maxHeight] = s.s[maxHeight-1]
		maxHeight++
	}

	// fill up the `nexts` field of k
	for i := 0; i < newHeight; i++ {
		prev := s.s[i].knot.nexts[i]
		k.nexts[i].knot = prev.knot
		k.nexts[i].skipped = byteCount + prev.skipped - s.s[i].skipped
		k.nexts[i].skippedRunes = runeCount + prev.skippedRunes - s.s[i].skippedRunes

		s.s[i].knot.nexts[i].knot = k
		s.s[i].knot.nexts[i].skipped = s.s[i].skipped
		s.s[i].knot.nexts[i].skippedRunes = s.s[i].skippedRunes

		// move search to end of newly inserted node
		s.s[i].knot = k
		s.s[i].skipped = byteCount
		s.s[i].skippedRunes = runeCount
	}

	for i := newHeight; i < maxHeight; i++ {
		s.s[i].knot.nexts[i].skipped += byteCount
		s.s[i].knot.nexts[i].skippedRunes += runeCount
		s.s[i].skipped += byteCount
		s.s[i].skippedRunes += runeCount
	}
	s.r.size += byteCount
	s.r.runes += runeCount
}

// find is the generic skip list finding function. It returns the offsets and skipped bytes.
func (s *skiplist) find(point int) (retVal *knot, offsetBytes, skippedBytes int, err error) {
	if point > s.r.runes {
		return nil, -1, -1, errors.New("Index out of bounds")
	}

	k := &s.r.Head
	height := k.height - 1
	offset := point

	for {
		var skip int
		skip = k.nexts[height].skippedRunes
		if offset > skip {
			// go right
			offset -= skip
			if k.nexts[height].knot == nil {
				break
			}
			skippedBytes += k.nexts[height].skipped
			k = k.nexts[height].knot
		} else {
			// go down
			s.s[height].skippedRunes = offset
			s.s[height].skipped = skippedBytes
			s.s[height].knot = k
			if height == 0 {
				break
			}
			height--
		}
	}
	offsetBytes = byteOffset(k.data[:], offset)
	return k, offsetBytes, skippedBytes, nil
}

// find2 is a method that finds blocks for insertion and deletion. No counts for byteoffsets required.
func (s *skiplist) find2(point int) (retVal *knot, err error) {
	if point > s.r.runes {
		return nil, errors.New("Index out of bounds")
	}

	k := &s.r.Head
	height := k.height - 1
	offset := point

	for {
		var skip int
		skip = k.nexts[height].skippedRunes
		if offset > skip {
			// go right
			offset -= skip
			if k.nexts[height].knot == nil {
				break
			}
			k = k.nexts[height].knot
		} else {
			// go down
			s.s[height].skippedRunes = offset
			s.s[height].knot = k
			if height == 0 {
				break
			}
			height--
		}
	}
	return k, nil
}

func (s *skiplist) updateOffsets(bytecount, runecount int) {
	for i := 0; i < s.r.Head.height; i++ {
		s.s[i].knot.nexts[i].skipped += bytecount
		s.s[i].knot.nexts[i].skippedRunes += runecount
	}
}

func (s *skiplist) insert(k *knot, data []byte) error {
	offset := s.s[0].skippedRunes
	var offsetBytes int
	if offset > 0 {
		offsetBytes = byteOffset(k.data[:], offset)
	}

	byteCount := len(data)

	// can insert?
	canInsert := k.used+byteCount <= BucketSize
	if !canInsert && offsetBytes == k.used {
		next := k.nexts[0].knot
		if next != nil && next.used+byteCount < BucketSize {
			offset = 0
			offsetBytes = 0
			for i := 0; i < next.height; i++ {
				s.s[i].knot = next
			}
			k = next
			canInsert = true
		}
	}
	runeCount := utf8.RuneCount(data)

	if canInsert {
		// move shit
		if byteCount < k.used {
			copy(k.data[offset+byteCount:], k.data[offset:])
		}
		copy(k.data[offset:offset+byteCount], data)
		k.used += byteCount
		s.r.size += byteCount
		s.r.runes += runeCount
		// update the rest of the search tree
		s.updateOffsets(byteCount, runeCount)
	} else {
		// we'll need to add at least Knot to the rope

		// we'll need to remove the end of the current node's data if this is not at the end of the current node
		endBytes := k.used - offsetBytes
		var endRunes int
		if endBytes > 0 {
			k.used = offsetBytes
			endRunes = k.nexts[0].skippedRunes - offset
			s.updateOffsets(-endBytes, -endRunes)
			s.r.size -= endBytes
			s.r.runes -= endRunes
		}

		// insert new Knots containing new data
		var dataOffset int
		for dataOffset < len(data) {
			var newBytes, newRunes int
			for dataOffset+newBytes < len(data) {
				_, width := utf8.DecodeRune(data[dataOffset+newBytes:])
				if newBytes+width > BucketSize {
					break
				}
				newBytes += width
				newRunes++
			}
			// create new Knot
			s.newKnot(data[dataOffset:dataOffset+newBytes], newRunes)
			dataOffset += newBytes
		}

		// if we removed the end, it's time to add it back
		if endBytes > 0 {
			s.newKnot(k.data[offsetBytes:offsetBytes+endBytes], endRunes)
		}
	}
	return nil
}

func (s *skiplist) del(k *knot, n int) {
	s.r.runes -= n
	offset := s.s[0].skippedRunes
	var i int
	for n > 0 {
		if k == nil {
			break
		}
		if offset == k.nexts[0].skippedRunes {
			// end found. skip to the start of the next node
			k = s.s[0].knot.nexts[0].knot
			offset = 0
		}
		size := k.nexts[0].skippedRunes
		removed := min(n, size-offset)

		if removed < size || k == &s.r.Head {
			leading := byteOffset(k.data[:], offset)
			removedBytes := byteOffset(k.data[leading:], removed)
			if trailing := k.used - leading - removedBytes; trailing > 0 {
				copy(k.data[offset:], k.data[offset+removedBytes:offset+removedBytes+trailing])
			}
			k.used -= removedBytes
			s.r.size -= removedBytes

			for i = 0; i < k.height; i++ {
				k.nexts[i].skippedRunes -= removed
			}
		} else {
			for i = 0; i < k.height; i++ {
				s.s[i].knot.nexts[i].knot = k.nexts[i].knot
				s.s[i].knot.nexts[i].skippedRunes += k.nexts[i].skippedRunes - removed
			}
			s.r.size -= k.used
			k = k.nexts[0].knot

		}
		for ; i < s.r.Head.height; i++ {
			s.s[i].knot.nexts[i].skippedRunes -= removed
		}
		n -= removed
	}
}
