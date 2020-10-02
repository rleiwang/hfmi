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

package hfmi

import "io"

// Succinct defines Rank/Select succinct data structure
type Succinct interface {
	// Access returns byte and its rank at p-th position, p is zero based offset
	Access(p uint) (a byte, r uint, ok bool)

	// Select returns the position p of r-th ranked a, p is zero based offset
	Select(a byte, r uint) (p uint, ok bool)

	// Rank returns the rank, r-th of byte at the position p, p is zero based offset
	Rank(a byte, p uint) (r uint, ok bool)

	// Bytes returns bit vector of this succinct data structure
	Bytes() []byte

	//
	Dictionary() []byte
}

type FMI interface {
	Succinct

	// Locate returns the bucket byte and its rank at p-th position, p is zero based offset
	Locate(uint) (byte, uint, bool)

	// Count the number of pattern occurrence
	Count(string) uint

	// Search search pattern, return range in BWT (s, e]
	Search(string) ([]uint, bool)

	// Size return the size of header and body bit vector
	Size() (int, int)

	// Len return original text len
	Len() uint

	// Histogram return [256]uint
	Histogram() []uint

	// CharsInBound return all chars between bound [start, end]
	CharsInBound(uint, uint) []byte

	// GetBound return (start, end] range bound
	GetBound(byte) (uint, uint, bool)

	//
	Restore(io.Writer) bool

	// ForwardExtractToChar return []byte, position, ok, walking BWT forward from p, until found b
	ForwardExtractToChar(uint, byte) ([]byte, uint, bool)

	// BackwardExtractToChar return []byte, position, ok, walking BWT backward from p, until found b
	BackwardExtractToChar(uint, byte) ([]byte, uint, bool)

	// BackwardJumpToChar return position, ok, walking BWT backward from p, until found b
	BackwardJumpToChar(uint, byte) (uint, bool)

	// ExtractFields, returns all fields where p falls in.
	// sep -> separator, p -> position of BWT, fc -> field count
	ExtractFields(sep byte, p uint, fc uint) ([][]byte, bool)

	ExtractAllFields(sep byte, fc uint) ([][][]byte, bool)

	// ExtractRange, return []byte between from and to or 0/1 byte, which ever comes first
	ExtractRange(from, to uint) ([]byte, bool)
}
