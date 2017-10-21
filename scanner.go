package skiprope

import (
	"errors"
	"io"
	"unicode/utf8"
)

var ErrSOF = errors.New("Cannot unread on SOF")

// Scanner is a linear scanner over a *Rope which returns runes.
type Scanner struct {
	*Rope
	k         *knot // current
	offset    int
	readBytes int // how many bytes has been read

	// lastSize is used in implementing UnreadRune()
	lastSize int
	prevK    *knot
}

// NewScanner creates a new scanner.
func NewScanner(r *Rope) *Scanner {
	sl := skiplist{r: r}
	s := &Scanner{Rope: r}
	var err error
	if s.k, s.offset, _, err = sl.find(0); err != nil {
		panic(err)
	}

	// first block may not be used
	for s.k.used-s.offset <= 0 {
		// go to next block
		s.prevK = s.k
		s.k = s.k.nexts[0].knot
	}
	return s
}

// Len returns the number of bytes unread
func (s *Scanner) Len() int { return s.size - s.readBytes }

// Read implements io.Reader. It reads up to len(p) bytes into p. The Scanner is a stateful reader,
// meaning it will keep track of how many bytes has been read.
func (s *Scanner) Read(p []byte) (n int, err error) {
	if s.readBytes >= s.size || s.k == nil {
		return 0, io.EOF
	}

	l := len(p)
	used := s.k.used - s.offset
	n = copy(p, s.k.data[s.offset:s.k.used])
	if l <= used {
		s.offset += n
		s.readBytes += n
		return
	}

	// TODO: recursive reading is simple to understand but performance is ??? .
	remainder := l - n
	if remainder > 0 {
		s.offset = 0

		var n2 int
		s.k = s.k.nexts[0].knot
		if n2, err = s.Read(p[remainder:]); err != nil && err != io.EOF {
			return n, err
		}
		err = nil
		n += n2
	}
	return
}

// ReadByte  implements io.ByteReader
func (s *Scanner) ReadByte() (byte, error) {
	if s.readBytes >= s.size || s.k == nil {
		return 0, io.EOF
	}
	retVal := s.k.data[s.offset]
	s.offset++
	s.readBytes++
	if s.offset >= s.k.used {
		s.prevK = s.k
		s.k = s.k.nexts[0].knot
		s.offset = 0
	}
	return retVal, nil
}

// ReadRune implements io.RuneReader. It reads a single UTF8 encoded character and then returns its size in bytes
func (s *Scanner) ReadRune() (rune, int, error) {
	if s.k == nil || s.readBytes >= s.size {
		return -1, -1, io.EOF
	}

	r, size := utf8.DecodeRune(s.k.data[s.offset:s.k.used])
	s.lastSize = size
	s.offset += size
	s.readBytes += size
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
	s.readBytes -= s.lastSize
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
