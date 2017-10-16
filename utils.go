package skiprope

import (
	"math/rand"
	"time"
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
