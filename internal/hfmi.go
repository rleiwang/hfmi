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

package internal

import "sync/atomic"

const (
	SZ = 256 // block must be N * 64
)

var (
	csz      uint64
	segments []byte
	counter  uint64
)

type SDS interface {
	Access(uint, []byte) (byte, uint)
	Rank(byte, uint, []byte) uint
}

type Encoder func([]byte, []byte, byte, []byte, []uint16) uint

func InitSegmentCache(sz uint64) {
	segments = make([]byte, SZ*sz)
	csz = sz
}

func GetFreeSegment() []byte {
	n := (atomic.AddUint64(&counter, 1) % csz) * SZ
	bv := segments[n : n+SZ]
	bv[0] = 0
	for i := 1; i <= 128; i <<= 1 {
		copy(bv[i:], bv[:i])
	}

	return bv
}

func CalcBlockHistogram(data []byte) ([]byte, []uint16, byte, uint) {
	chars, hist, prev, runs := [256]byte{}, [256]uint16{}, data[0], uint(1)
	hist[prev]++
	for _, c := range data[1:] {
		hist[c]++
		if prev != c {
			runs++
			prev = c
		}
	}

	e, mfc := 0, uint16(0)
	for i, h := range hist {
		if h > 0 {
			hist[e] = h
			chars[e] = byte(i)
			if h > mfc {
				mfc = h
				prev = byte(i)
			}
			e++
		}
	}

	return chars[:e], hist[:e], prev, runs
}
