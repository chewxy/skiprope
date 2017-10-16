package skiprope

import (
	"github.com/pkg/errors"
)

// skiplist is a data structure for searching... it's the "skiplist" part of things
type skiplist struct {
	r *Rope
	s [MaxHeight]skipknot
}

// newKnot will accept a []rune of BucketSize or less
func (s *skiplist) newKnot(data []rune) {
	maxHeight := s.r.Head.height
	newHeight := randInt()

	k := newKnot(newHeight)
	k.used = len(data)
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
		k.nexts[i].skipped = len(data) + prev.skipped - s.s[i].skipped

		s.s[i].knot.nexts[i].knot = k
		s.s[i].knot.nexts[i].skipped = s.s[i].skipped

		// move search to end of newly inserted node
		s.s[i].knot = k
		s.s[i].skipped = len(data)
	}

	for i := newHeight; i < maxHeight; i++ {
		s.s[i].knot.nexts[i].skipped += len(data)

		s.s[i].skipped += len(data)
	}
	s.r.size += len(data)
}

func (s *skiplist) find(point int) (retVal *knot, offset, lines int, err error) {
	if point > s.r.size {
		return nil, -1, -1, errors.Errorf("Index out of bounds")
	}

	k := &s.r.Head
	height := k.height - 1
	offset = point

	var skip int
	for {
		skip = k.nexts[height].skipped
		// lines = k.nexts[height].skippedLines
		if offset > skip {
			// go right
			offset -= skip
			if k.nexts[height].knot == nil {
				break
			}
			k = k.nexts[height].knot
		} else {
			// go down
			s.s[height].skipped = offset
			s.s[height].knot = k
			if height == 0 {
				break
			}
			height--
		}
	}
	return k, offset, lines, nil
}

func (s *skiplist) updateOffsets(count int) {
	for i := 0; i < s.r.Head.height; i++ {
		s.s[i].knot.nexts[i].skipped += count
	}
}

func (s *skiplist) insert(k *knot, data []rune) error {
	offset := s.s[0].skipped
	// if offset > 0 {
	// 	// check
	// 	if len(k.nexts) < 0 {
	// 		return errors.Errorf("Not enoug nexts")
	// 	}
	// 	if offset > k.nexts[0].skipped {
	// 		return errors.Errorf("Index out of boudns")
	// 	}
	// }

	// can insert?
	canInsert := k.used+len(data) <= BucketSize
	if !canInsert && offset == k.used {
		next := k.nexts[0].knot
		if next != nil && next.used+len(data) < BucketSize {
			offset = 0
			for i := 0; i < next.height; i++ {
				s.s[i].knot = next
			}
			k = next
			canInsert = true
		}
	}

	if canInsert {
		// move shit
		if len(data) < k.used {
			copy(k.data[offset+len(data):], k.data[offset:])
		}
		copy(k.data[offset:offset+len(data)], data)
		k.used += len(data)
		s.r.size += len(data)
		// update the rest of the search tree
		s.updateOffsets(len(data))
	} else {
		// we'll need to add at least Knot to the rope

		// we'll need to remove the end of the current node's data if this is not at the end of the current node
		endCount := k.used - offset
		if endCount > 0 {
			k.used = offset
			endCount = k.nexts[0].skipped - offset
			s.updateOffsets(-endCount)

			s.r.size -= endCount
		}

		// insert new Knots containing new data
		var dataOffset int
		for dataOffset < len(data) {
			var newCount int
			for dataOffset+newCount < len(data) {
				if newCount+1 > BucketSize {
					break
				}
				newCount++
			}
			// create new Knot
			s.newKnot(data[dataOffset : dataOffset+newCount])
			dataOffset += newCount
		}

		// if we removed the end, it's time to add it back
		if endCount > 0 {
			s.newKnot(k.data[offset:])
		}
	}
	return nil
}

func (s *skiplist) del(k *knot, n int) {
	s.r.size -= n
	offset := s.s[0].skipped
	var i int
	for n > 0 {
		if k == nil {
			break
		}
		if offset == k.nexts[0].skipped {
			// end found. skip to the start of the next node
			k = s.s[0].knot.nexts[0].knot
			offset = 0
		}
		size := k.nexts[0].skipped
		removed := min(n, size-offset)

		if removed < size || k == &s.r.Head {
			if trailing := k.used - offset - removed; trailing > 0 {
				copy(k.data[offset:], k.data[offset+removed:offset+removed+trailing])
			}
			k.used -= removed

			for i = 0; i < k.height; i++ {
				k.nexts[i].skipped -= removed
			}
		} else {
			for i = 0; i < k.height; i++ {
				s.s[i].knot.nexts[i].knot = k.nexts[i].knot
				s.s[i].knot.nexts[i].skipped += k.nexts[i].skipped - removed
			}
			k = k.nexts[0].knot

		}
		for ; i < s.r.Head.height; i++ {
			s.s[i].knot.nexts[i].skipped -= removed
		}
		n -= removed
	}
}
