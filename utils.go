package skiprope

import (
	"math/rand"
	"time"
	"unicode/utf8"
	// "log"
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

func randInt() (retVal int) {
	retVal = 1
	for retVal < MaxHeight-1 && src.Intn(100) < Bias {
		retVal++
	}
	return retVal
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clamp(a, minVal, maxVal int) int {
	return max(minVal, min(maxVal, a))
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
