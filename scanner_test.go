package skiprope

import (
	"fmt"
	"io"
	"testing"
)

func ExampleScanner() {
	r := New()
	r.Insert(0, "Hello 世界")

	s := NewScanner(r)
	for r, _, err := s.ReadRune(); err == nil; r, _, err = s.ReadRune() {
		fmt.Printf("%q\n", r)
	}
	// Output:
	// 'H'
	// 'e'
	// 'l'
	// 'l'
	// 'o'
	// ' '
	// '世'
	// '界'
}

func ExampleScanner_long() {
	r := New()
	r.Insert(0, "Hello 世界. This is a longer sentence that spans multiple blocks of skip nodes.")

	s := NewScanner(r)
	for r, _, err := s.ReadRune(); err == nil; r, _, err = s.ReadRune() {
		fmt.Printf("%q\n", r)
	}
	// Output:
	// 'H'
	// 'e'
	// 'l'
	// 'l'
	// 'o'
	// ' '
	// '世'
	// '界'
	// '.'
	// ' '
	// 'T'
	// 'h'
	// 'i'
	// 's'
	// ' '
	// 'i'
	// 's'
	// ' '
	// 'a'
	// ' '
	// 'l'
	// 'o'
	// 'n'
	// 'g'
	// 'e'
	// 'r'
	// ' '
	// 's'
	// 'e'
	// 'n'
	// 't'
	// 'e'
	// 'n'
	// 'c'
	// 'e'
	// ' '
	// 't'
	// 'h'
	// 'a'
	// 't'
	// ' '
	// 's'
	// 'p'
	// 'a'
	// 'n'
	// 's'
	// ' '
	// 'm'
	// 'u'
	// 'l'
	// 't'
	// 'i'
	// 'p'
	// 'l'
	// 'e'
	// ' '
	// 'b'
	// 'l'
	// 'o'
	// 'c'
	// 'k'
	// 's'
	// ' '
	// 'o'
	// 'f'
	// ' '
	// 's'
	// 'k'
	// 'i'
	// 'p'
	// ' '
	// 'n'
	// 'o'
	// 'd'
	// 'e'
	// 's'
	// '.'
}

func ExampleScanner_UnreadRune() {
	r := New()
	r.Insert(0, "Hello 世界")

	s := NewScanner(r)
	s.ReadRune()
	s.UnreadRune()
	char, _, _ := s.ReadRune()
	fmt.Printf("Read-Unread-Read: %q", char)
	// Output:
	// Read-Unread-Read: 'H'
}

func TestScanner_UnreadRune(t *testing.T) {
	r := New()
	r.Insert(0, "Hello 世界")
	s := NewScanner(r)
	var err error
	if err = s.UnreadRune(); err == nil {
		t.Error("Error was expected when trying to unread rune of 0")
	}

	r.Insert(0, "Some other long string")
	for _, _, err = s.ReadRune(); err == nil; _, _, err = s.ReadRune() {

	}
	if err != io.EOF {
		t.Error(err)
	}
	for err = s.UnreadRune(); err == nil; err = s.UnreadRune() {
	}
	if err != ErrSOF {
		t.Error(err)
	}
}

func TestScanner_Read(t *testing.T) {
	r := New()
	r.Insert(0, "This is a  long string that is meant to span multiple *knots")

	p := make([]byte, 5)
	s := NewScanner(r)
	n, err := s.Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Error("Expected a read of 5 bytes")
	}
	if string(p) != "This " {
		t.Error("First read failed")
	}
	n, err = s.Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(p) != "is a " {
		t.Error("2nd read failed")
	}

	// Longer stuff
	s = NewScanner(r)
	p = make([]byte, 1024)
	if n, err = s.Read(p); err != nil {
		t.Fatal(err)
	}
	if string(p[:n]) != "This is a  long string that is meant to span multiple *knots" {
		t.Error("Long read failed")
	}
}

// ReadByte is affected by other method calls, as demonstrated by caling ReadRune
func ExampleScanner_ReadByte() {
	r := New()
	r.Insert(0, "Hello World")

	s := NewScanner(r)
	fmt.Printf("String Size: %d.\n", len("Hello World"))
	s.ReadByte() // H
	s.ReadRune() // E
	fmt.Printf("Bytes Unread: %d\n", s.Len())
	for b, err := s.ReadByte(); err == nil; b, err = s.ReadByte() {
		fmt.Printf("%q\n", b)
	}

	// Output:
	// String Size: 11.
	// Bytes Unread: 9
	// 'l'
	// 'l'
	// 'o'
	// ' '
	// 'W'
	// 'o'
	// 'r'
	// 'l'
	// 'd'
}
