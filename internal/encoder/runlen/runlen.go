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

package runlen

import (
	"sync"

	"github.com/rleiwang/hfmi/internal"
)

// RUN LENGTH ENCODER
// ┌─┬────────┐
// │0│ Char 1 │
// ├─┼────────┤
// │1│ Length │
// ├─┼────────┤
// │2│ Char 2 │
// ├─┼────────┤
// │3│ Length │
// └─┴────────┘

type runlen struct {
}

var (
	RunLength internal.SDS

	zeros [256]uint

	pool = sync.Pool{
		New: func() interface{} { return &[256]uint{} },
	}
)

func init() {
	RunLength = &runlen{}
}

func (*runlen) Access(p uint, bv []byte) (byte, uint) {
	b, ptr, ranks := byte(0), uint(0), pool.Get().(*[256]uint)
	copy(ranks[:], zeros[:])
	for i := 0; i < len(bv); i++ {
		// character
		b = bv[i]
		i++
		if ptr+uint(bv[i]) > p {
			// the current run is over the p
			break
		}
		// log ranks
		ranks[b] += uint(bv[i])
		// ptr += the length of current run
		ptr += uint(bv[i])
	}

	pool.Put(ranks)

	// ranks = current pos (p) - offset of current run (last) + previous rank of b
	return b, p - ptr + ranks[b] + 1
}

func (*runlen) Rank(a byte, p uint, bv []byte) uint {
	offset, last, r := uint(0), uint(0), uint(0)
	for i := 0; i < len(bv); i++ {
		// skip character
		b := bv[i]
		i++
		// offsets += the length of current run
		offset += uint(bv[i])
		if offset > p {
			// ranks = current pos (p) - offset of current run (last) + previous rank of b
			if a == b {
				return p - last + r + 1
			}
			return r
		}
		if b == a {
			r += uint(bv[i])
		}

		last = offset
	}
	return r
}

func Encode(dst, src []byte, mfc byte, chars []byte, hist []uint16) uint {
	runs, prev, offset := byte(1), src[0], uint(0)
	for _, b := range src[1:] {
		if prev != b {
			dst[offset], dst[offset+1] = prev, runs
			runs, prev, offset = 1, b, offset+2
		} else {
			runs++
		}
	}
	dst[offset], dst[offset+1] = prev, runs
	return offset + 2
}

func CompSZ(chars []byte, hist []uint16, runs uint) uint {
	// two bytes per run
	return 2 * runs
}
