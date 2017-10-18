# SkipRope [![GoDoc](https://godoc.org/github.com/chewxy/skiprope?status.svg)](https://godoc.org/github.com/chewxy/skiprope) [![Build Status](https://travis-ci.org/chewxy/skiprope.svg?branch=master)](https://travis-ci.org/chewxy/skiprope) [![Coverage Status](https://coveralls.io/repos/github/chewxy/skiprope/badge.svg?branch=master)](https://coveralls.io/github/chewxy/skiprope?branch=master) #

package `skiprope` is an implementation of the [rope](https://en.wikipedia.org/wiki/Rope_%28data_structure%29) data structure. It's not strictly a rope as per se. It's not built on top of a binary tree like most rope data structures are. 

Instead it's built on top of a [skip list](https://en.wikipedia.org/wiki/Skip_list).  This makes it more akin to a piece table than a true rope. 

This library is quite complete as it is (it has all the functions and I need for my projects). However, any addition, corrections etc are welcome. Please see CONTRIBUTING.md for more information on how to contribute. 

# Installing #

This package is go-gettable: `go get -u github.com/chewxy/skiprope`. 

This package is versioned with SemVer 2.0, and does not have any external dependencies except for testing (of which the fantastic [`assert`](https://github.com/stretchr/testify) package is used).

# Usage #

```
r := skiprope.New()
if err := r.Insert(0, "Hello World! Hello, 世界"); err != nil {
	// handle error
}
if err  := r.Insert(5, "there "); err != nil {
	// handle error
}
fmt.Println(r.String()) // "Hello there World! Hello, 世界"
```

# FAQ #

## How do I keep track of row, col information? ##

The best way is to use a linebuffer data structure in conjunction with the `Rope`:

```
type Foo struct{
	*skiprope.Rope
	linebuf []int
}
```

The `linebuf` is essentially a list of offsets to a newline character. The main reason why I didn't add a `linebuf` slice into the `*Rope` data structure is because while it's a common thing to do, for a couple of my projects I didn't need it. The nature of Go's composable data structure and programs make it quite easy to add these things. A pro-tip is to extend the `Insert`, `InsertRunes` and `EraseAt` methods.

I do not think it to be wise to add it to the core data structure.

## When is `*Rope` going to implement `io.Writer` and `io.Reader`? ##

The main reason why I didn't do it was mostly because I didn't need it. However, I've been asked about this before. I personally don't have the bandwidth to do it.

Please send a pull request :)

# Benchmarks #

There is a benchmark mini-program that is not required for the running, but here are the results:

```
go test -run=^$ -bench=. -benchmem -cpuprofile=test.prof
BenchmarkNaiveRandomInsert-8   	   50000	     37774 ns/op	      15 B/op	       0 allocs/op
BenchmarkRopeRandomInsert-8    	 3000000	       407 ns/op	      27 B/op	       0 allocs/op
BenchmarkERopeRandomInsert-8   	 1000000	      2143 ns/op	    1161 B/op	      25 allocs/op
PASS
```

# History, Development and Acknowledgements #
This package started its life as a textbook binary-tree data structure for another personal project of mine. Over time, I decided to add more and more optimizations to it. The original design had a split at every new-line character. 

I started by moving more and more things off the heap, and onto the stack. As I wondered how to incorporate a search structure using a skiplist, I stumbled onto a well developed library for C, [librope](https://github.com/josephg/librope).

It had everything I had wanted: a rope-like structure, using skiplists to find nodes, minimal allocations on heap, and a solution to my problem wrt keeping the skiplists off heap. The solution turned out to be to be the `skiplist` data structure, without a pointer. So I ended up adapting most of Joseph Gentle's algorithm. So this library owes most of its work to Joseph Gentle. 

Hours before releasing this library, I had a consult by Egon Elbre, who gave good advice on whether just sticking with `[]rune` was a good idea. He managed to convince me that it isn't, so the first pull request was made to update this library to deal with `[]byte` instead. As a result, memory use went down 40B/op at the cost of an increase of about 30 ns/op. The number can be further shaved down with better optimizations.