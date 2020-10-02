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

// Code a node in Huffman Tree
type Code struct {
	Freq   uint   // frequence of the symbol in text
	Len    byte   // length of the prefix code
	Prefix uint16 // code of the prefix encoding
}

// internal code
type code struct {
	Code
	s int16 // the actual symbol in text
	l *code // left node of the Huffman tree
	r *code // right node of the Huffman tree
}

// -- min heap interface by word frequency --

type minHeapByFreq []*code

func (m minHeapByFreq) Len() int {
	return len(m)
}

func (m minHeapByFreq) Less(l, r int) bool {
	// Pop to return lowest freq
	if m[l].Freq == m[r].Freq {
		// internal nodes s < 0, return internal node first. internal -> leaf
		// leaf nodes in lexicographic ascending order
		return m[l].s < m[r].s
	}
	// freq low -> high
	return m[l].Freq < m[r].Freq
}

func (m minHeapByFreq) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *minHeapByFreq) Push(n interface{}) {
	*m = append(*m, n.(*code))
}

func (m *minHeapByFreq) Pop() interface{} {
	n := len(*m)
	item := (*m)[n-1]
	*m = (*m)[:n-1]
	return item
}

// -- end heap interface --
