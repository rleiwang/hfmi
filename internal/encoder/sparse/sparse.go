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

package sparse

import (
	"sync"

	"github.com/rleiwang/hfmi/internal"
)

// Sparse encoder keeps absolute offset of sparse character
// [cccccccBccccccAcccccc],  A and B are sparse character
// ┌─┬────┐
// │0│ B  │
// ├─┼────┤
// │1│ 7  │
// ├─┼────┤
// │2│ A  │
// ├─┼────┤
// │3│ 14 │
// └─┴────┘

type sparse struct {
	mfc byte
}

var (
	zeros [256]uint

	pool = sync.Pool{
		New: func() interface{} { return &[256]uint{} },
	}
)

func New(mfc byte) internal.SDS {
	return &sparse{mfc}
}

func (s *sparse) Access(p uint, bv []byte) (byte, uint) {
	// find out the most frequented character
	b, i, ranks := byte(0), 0, pool.Get().(*[256]uint)
	copy(ranks[:], zeros[:])

	// iterate through sparse character
	for i = 0; i < len(bv); i++ {
		// skip character
		b = bv[i]
		i++
		if uint(bv[i]) == p {
			pool.Put(ranks)
			return b, ranks[b] + 1
		} else if uint(bv[i]) > p {
			// offset is over, T[p] is the frequent character
			break
		}
		ranks[b]++
	}

	pool.Put(ranks)

	// ranks is pos - offset excludes sparse character
	return s.mfc, p - uint(i)>>1 + 1
}

func (s *sparse) Rank(a byte, p uint, bv []byte) uint {
	// cnt -> sparse char counts
	cnt, r, b := uint(0), uint(0), byte(0)

	// iterate through sparse character
	for i := 0; i < len(bv); i++ {
		// skip character
		b = bv[i]
		i++
		if uint(bv[i]) > p {
			break
		}
		if b == a {
			r++
		}
		cnt++
	}

	if a == s.mfc {
		// a is most frequent char, ranks = p - number_of_sparse chars
		return p - cnt + 1
	}

	return r
}

func Encode(dst, src []byte, mfc byte, chars []byte, hist []uint16) uint {
	i := 0
	for j, c := range src {
		if mfc != c {
			dst[i] = c
			i++
			dst[i] = byte(j)
			i++
		}
	}

	return uint(i)
}

func CompSZ(chars []byte, hist []uint16, runs uint) uint {
	//  hist are in freq desc order
	s := uint(0)
	for _, h := range hist[1:] {
		s += uint(h) * 2
	}
	return s
}
