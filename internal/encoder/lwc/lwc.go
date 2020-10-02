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

package lwc

import (
	"github.com/rleiwang/hfmi/internal"
)

type lwc struct {
}

var (
	LWC    internal.SDS
	weight [256]byte
)

func init() {
	LWC = &lwc{}
	weight[0] = 1
}

func Expand(src, chars []byte) []byte {
	dst, sz := internal.GetFreeSegment(), len(src)
	if len(chars) < 3 {
		_ = dst[internal.SZ-1]
		for i := 0; i < sz; i++ {
			copy(dst[i*8:], single[src[i]])
		}
		sz *= 8
	} else if len(chars) < 5 {
		_ = dst[internal.SZ-1]
		for i := 0; i < sz; i++ {
			copy(dst[i*4:], half[src[i]])
		}
		sz *= 4
	} else if len(chars) < 17 {
		_ = dst[internal.SZ-1]
		for i := 0; i < sz; i++ {
			copy(dst[i*2:], nibble[src[i]])
		}
		sz *= 2
	} else {
		copy(dst, src)
	}

	for i, c := range dst[:sz] {
		dst[i] = chars[c]
	}

	return dst
}

func (l *lwc) Access(p uint, bv []byte) (byte, uint) {
	a, r := bv[p], uint(0)
	for _, b := range bv[:p+1] {
		r += uint(weight[a^b])
	}
	return a, r
}

func (l *lwc) Rank(a byte, p uint, bv []byte) uint {
	r := uint(0)
	for _, b := range bv[:p+1] {
		r += uint(weight[a^b])
	}
	return r
}

func Encode(dst, src []byte, mfc byte, chars []byte, hist []uint16) uint {
	convert := internal.GetFreeSegment()
	for i, c := range chars {
		convert[c] = byte(i)
	}

	cnt := uint(len(src))
	if len(chars) < 3 {
		for len(src) >= 8 {
			dst[0] = convert[src[0]]&0x1 |
				(convert[src[1]]&0x1)<<1 |
				(convert[src[2]]&0x1)<<2 |
				(convert[src[3]]&0x1)<<3 |
				(convert[src[4]]&0x1)<<4 |
				(convert[src[5]]&0x1)<<5 |
				(convert[src[6]]&0x1)<<6 |
				(convert[src[7]]&0x1)<<7
			dst, src = dst[1:], src[8:]
		}
		if len(src) > 0 {
			dst[0] = 0
			for i, c := range src {
				dst[0] |= (convert[c] & 0x1) << i
			}
		}
		cnt = (cnt + 7) / 8
	} else if len(chars) < 5 {
		for len(src) >= 4 {
			dst[0] = convert[src[0]]&0x3 |
				(convert[src[1]]&0x3)<<2 |
				(convert[src[2]]&0x3)<<4 |
				(convert[src[3]]&0x3)<<6
			dst, src = dst[1:], src[4:]
		}
		if len(src) > 0 {
			dst[0] = 0
			for i, c := range src {
				dst[0] |= (convert[c] & 0x3) << (i * 2)
			}
		}
		cnt = (cnt + 3) / 4
	} else if len(chars) < 17 {
		for len(src) >= 2 {
			dst[0] = convert[src[0]]&0xF |
				(convert[src[1]]&0xF)<<4
			dst, src = dst[1:], src[2:]
		}
		if len(src) > 0 {
			dst[0] = convert[src[0]] & 0xF
		}
		cnt = (cnt + 1) / 2
	} else {
		for i, c := range src {
			dst[i] = convert[c]
		}
	}

	return cnt
}

// CompSZ compute the size compressed array
func CompSZ(chars []byte, hist []uint16, runs uint) uint {
	if len(chars) < 3 {
		return internal.SZ >> 3
	} else if len(chars) < 5 {
		return internal.SZ >> 2
	} else if len(chars) < 17 {
		return internal.SZ >> 1
	}
	return internal.SZ
}
