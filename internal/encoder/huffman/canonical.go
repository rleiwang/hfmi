/*
 * Copyright 2020 Rock Lei Wang
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Package parser declares an expression parser with support for macro
 * expansion.
 */

package huffman

import (
	"container/heap"
	"sort"
)

// Sizing sizing huffman code length
func CompSZ(chars []byte, hist []uint16) uint {
	// min heap by the freq desc
	totalBits, la := uint(0), len(chars)
	tmp, mn := createTmpCode(la)
	for _, n := range buildMinHeap(toArray(chars, hist, tmp[:la], mn[:la]), tmp[la:], mn[la:]) {
		for l := byte(0); l < n.Len; l++ {
			totalBits += uint(n.Freq)
		}
	}

	return (totalBits + 7) / 8
}

// CanonicalCode returns canonical Huffman code, and depth of the tree
func CanonicalCode(chars []byte, hist []uint16) ([256]*Code, byte) {
	// assign canonical code.
	// init {1, 0}
	// if no node with code length one
	// generate new code {11, 01, 10, 00}
	// one node with length two, assign {11}, remaining {01, 10, 00}
	// generate new code {011, 101, 001, 010, 100, 000}
	// two nodes with length three, {011, 101}, remainning {001, 010, 100, 000}
	// generate new code {0011, 0101, 1001, 0001, 0010, 0100, 10000, 0000}
	// until the end

	// min heap by the freq desc
	la := len(chars)
	tmp, mn := createTmpCode(la)
	n := buildMinHeap(toArray(chars, hist, tmp[:la], mn[:la]), tmp[la:], mn[la:])
	ln := len(n)

	// number of levels equals to longest code length
	depth := n[ln-1].Len

	// iterate from level 1 to max levels
	code := []uint16{1, 0}
	for i, l := 0, byte(1); l <= depth; l++ {
		c := 0
		// tree is sorted by code length in ascending order
		for i < ln && l == n[i].Len {
			// assign code with the same length, take {11}, remaining {01, 10, 00}
			n[i].Prefix = code[c]
			i++
			c++
		}
		// generate new code from the remaining code set
		code = genCanonicalCode(code[c:])
	}

	var m [256]*Code
	for _, nn := range n {
		m[byte(nn.s)] = &nn.Code
	}

	return m, depth
}

func toArray(chars []byte, hist []uint16, n []code, m []*code) []*code {
	for i, a := range chars {
		m[i] = &n[i]
		m[i].Freq = uint(hist[i])
		m[i].Len = 1
		m[i].s = int16(a)
	}
	return m
}

// Build minHeap of canonical Huffman Tree
func buildMinHeap(m minHeapByFreq, tmp []code, n []*code) []*code {
	// min heap by symbol frequency
	heap.Init(&m)
	t := build(m, tmp, n)

	// sort by code length in ascending order
	sort.Slice(t, func(i, j int) bool {
		if t[i].Len == t[j].Len {
			// code length is the same
			if t[i].Freq == t[j].Freq {
				// freq is the same, asc alphabetic order
				return t[i].s < t[j].s
			}
			// desc freq order
			return t[i].Freq > t[j].Freq
		}
		// asc code length
		return t[i].Len < t[j].Len
	})

	return t
}

func l(b int) int {
	t := b
	for t&(t-1) > 0 {
		t = t & (t - 1)
	}
	if t != b {
		return t << 1
	}
	return b
}

func createTmpCode(a int) ([]code, []*code) {
	la := a + l(a)
	return make([]code, la, la), make([]*code, a*2)
}

func build(m minHeapByFreq, tmp []code, n []*code) []*code {
	internal := int16(-1)
	copy(n, m)

	for len(m) > 1 {
		// a new internal code, two children codes with two lowest frequency
		tn := &tmp[^internal]
		tn.l = heap.Pop(&m).(*code)
		tn.r = heap.Pop(&m).(*code)
		tn.s = internal
		tn.Freq = tn.l.Freq + tn.r.Freq
		heap.Push(&m, tn)
		internal--
	}

	calcCodeLen(m[0], 0)

	return n
}

// prefix length is the depth of Huffman Tree
func calcCodeLen(n *code, d byte) {
	if n.l == nil {
		n.Len = d
	} else {
		calcCodeLen(n.l, d+1)
		calcCodeLen(n.r, d+1)
	}
}

// generate new code {1, 0} -> {11, 01, 10, 00}
// {1, 0} -> {11, 01} append 1
// {1, 0} -> {10, 00} append 0
func genCanonicalCode(code []uint16) []uint16 {
	newCode := []uint16{}
	for _, j := range []uint16{1, 0} {
		for _, c := range code {
			newCode = append(newCode, (c<<1)|j)
		}
	}

	return newCode
}
