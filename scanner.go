package skiprope

import (
	"errors"
	"io"
	"unicode/utf8"
)

var ErrSOF = errors.New("Cannot unread on SOF")

// Scanner is a linear scanner over a *Rope which returns runes.
type Scanner struct {
	skiplist
	k      *knot // current
	offset int

	// lastSize is used in implementing UnreadRune()
	lastSize int
	prevK    *knot
}

// NewScanner creates a new scanner.
func NewScanner(r *Rope) *Scanner {
	s := &Scanner{
		skiplist: skiplist{r: r},
	}
	var err error
	if s.k, s.offset, _, err = s.find(0); err != nil {
		panic(err)
	}
	return s
}

// ReadRune implements io.RuneReader. It reads a single UTF8 encoded character and then returns its size in bytes
func (s *Scanner) ReadRune() (rune, int, error) {
	if s.k == nil {
		return -1, -1, io.EOF
	}
	if s.k.used-s.offset <= 0 {
		// go to next block
		s.prevK = s.k
		s.k = s.k.nexts[0].knot
	}
	r, size := utf8.DecodeRune(s.k.data[s.offset:s.k.used])
	s.lastSize = size
	s.offset += size
	if s.offset >= s.k.used {
		s.prevK = s.k
		s.k = s.k.nexts[0].knot
		s.offset = 0
	}
	return r, size, nil
}

// UnreadRune implements io.RuneScanner.
func (s *Scanner) UnreadRune() error {
	if s.offset == 0 && s.prevK == nil {
		return ErrSOF
	}
	if s.offset == 0 {
		s.k = s.prevK
		s.prevK = nil
		s.offset = s.k.used - s.lastSize
	} else {
		s.offset -= s.lastSize
		if s.offset < 0 {
			// TODO: this is unlikely to happen
			return io.ErrShortBuffer
		}
	}
	return nil
}
