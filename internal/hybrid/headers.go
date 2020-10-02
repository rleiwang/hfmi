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

package hybrid

import (
	"encoding/binary"

	"github.com/rleiwang/hfmi/internal"
	lwcenc "github.com/rleiwang/hfmi/internal/encoder/lwc"
	rulenc "github.com/rleiwang/hfmi/internal/encoder/runlen"
	sglenc "github.com/rleiwang/hfmi/internal/encoder/single"
	spsenc "github.com/rleiwang/hfmi/internal/encoder/sparse"
)

const (
	htp  = 5
	mask = byte(0x1F)
)

var (
	coefficient [256]byte
)

func init() {
	coefficient[0] = 1
}

// encodeHeader encode header as runlen, sparse and lwc
func encodeHeader(chars []byte, hist []uint16, e edt, bsz uint, header []byte) uint {
	if e == single {
		// +7+6+5+4+3+2+1+0+
		// |1|0|0|0|0|0|0|0| ◀──── meta
		// +-+-+-+-+-+-+-+-+
		// +-+-+-+-+-+-+-+-+
		// |0|0|0|0|0|0|0|0| ◀──── single char
		// +-+-+-+-+-+-+-+-+
		// +-+-+-+-+-+-+-+-+
		// |0|0|0|0|0|0|0|0| ◀──── freq 0 -> 256
		// +-+-+-+-+-+-+-+-+
		// note: for single char, there is no bv
		_ = header[2]
		header[0] = msb
		header[1] = chars[0]
		// note: if freq == 256, header[2] == zero
		header[2] = byte(hist[0])
		return 3
	}

	i := uint(1)
	// number of symbols in this block
	cnt := uint(len(chars))
	if cnt > 1<<htp {
		// +7+6+5+4+3+2+1+0+
		// |0|1|1|0|0|0|0|0| ◀──── meta
		// +-+-+-+-+-+-+-+-+
		// +-+-+-+-+-+-+-+-+
		// |0|0|1|0|0|0|0|0| ◀──── # of chars
		// +-+-+-+-+-+-+-+-+
		// contains 2 bytes and use msb to indicate that cnt > 2^5
		// first byte -> MSB (3 bits): type, LSB(5 bits): 0
		// second byte -> cnt (number of symbols)
		header[0], header[1] = byte(e)<<htp, byte(cnt)
		i = 2
	} else {
		// +7+6+5+4+3+2+1+0+
		// |1|0|1|0|1|1|1|1| ◀──── meta + # of chars
		// +-+-+-+-+-+-+-+-+
		// cnt <= 2^5, header contains 1 byte
		// one byte -> MSB (3 bits): type, LSB(5 bits): count
		// mask with 0001_1111
		// note: cnt == 2^5 results in 0_0000
		header[0] = msb | byte(cnt)&mask | byte(e)<<htp
	}

	// +7+6+5+4+3+2+1+0+
	// |0|0|1|0|1|1|1|1| ◀──── size of bv
	// +-+-+-+-+-+-+-+-+
	// size of the bv, note: size of bv == 0 iff len(bsz) = 256
	header[i] = byte(bsz)
	i++

	for j, k := range chars {
		// +-+-+-+-+-+-+-+-+
		// |0|0|0|0|1|0|1|0| ◀──── char
		// +-+-+-+-+-+-+-+-+
		// +-+-+-+-+-+-+-+-+
		// |0|0|0|0|1|0|0|0| ◀──── freq, must be less than 256
		// +-+-+-+-+-+-+-+-+
		header[i] = k
		i++
		header[i] = byte(hist[j] & 0xFF)
		i++
	}

	return i
}

func restoreHeader(h *hybrid) *hybrid {
	end, rank := uint(0), [256]uint{}

	count := binary.LittleEndian.Uint32(h.hdr[:4])
	h.m.char, h.m.hist = make([]byte, count), make([]uint16, count)
	h.m.bsz, h.m.bsds = make([]uint16, 0, 1+h.cnt/internal.SZ), make([]internal.SDS, 0, 1+h.cnt/internal.SZ)
	h.m.bbv = make([][]byte, 0, 1+h.cnt/internal.SZ)

	for i, j := 4, uint(0); i < len(h.hdr); i++ {
		cnt, t := uint(1), single
		if h.hdr[i]&msb != 0 {
			// msb is set, if single, 1000_0000
			if h.hdr[i] > msb {
				// msb is set, # of chars <= 2^5, 1010_0100
				t = edt(0x03 & (h.hdr[i] >> htp))
				cnt = uint(h.hdr[i] & mask)
				if cnt == 0 {
					cnt = 32
				}
			}
		} else {
			// 0100_0000, extract bit 5 and 6 to get edt
			t = edt(h.hdr[i] >> htp)
			i++
			// encoding scheme ensures # of chars in each block always be less than 256
			cnt = uint(h.hdr[i])
		}

		var bv []byte
		if cnt > 1 {
			beg := end
			i++
			end += uint(h.hdr[i])
			if end == beg {
				end += 256
			}
			bv = h.bv[beg:end]
		}

		next := j + cnt
		for s := j; s < next; s++ {
			i++
			c := h.hdr[i]
			i++
			freq := uint16(h.hdr[i])
			if freq == 0 {
				freq = 256
			}
			h.m.char[s], h.m.hist[s] = c, freq
			rank[c] += uint(freq)
		}

		switch t {
		case single:
			h.m.bsds = append(h.m.bsds, sglenc.New(h.m.char[j], h.m.hist[j]))
			// single char block has no body
			h.m.bbv = append(h.m.bbv, nil)
		case runlen:
			h.m.bsds = append(h.m.bsds, rulenc.RunLength)
			h.m.bbv = append(h.m.bbv, bv)
		case sparse:
			h.m.bsds = append(h.m.bsds, spsenc.New(findMostFreqChar(h.m.char[j:], h.m.hist[j:])))
			h.m.bbv = append(h.m.bbv, bv)
		case lwc:
			h.m.bsds = append(h.m.bsds, lwcenc.LWC)
			h.m.bbv = append(h.m.bbv, lwcenc.Expand(bv, h.m.char[j:next]))
		}

		h.m.bsz = append(h.m.bsz, uint16(cnt))
		j = next

		if len(h.m.bsz)%sbsz == 0 {
			h.m.super = append(h.m.super, super{offset: j})
			copy(h.m.super[len(h.m.super)-1].rank[:], rank[:])
		}
	}

	// count sentinel, the end of block
	offset, idx := rank[0], byte(1)
	h.m.eob = make([]pair, 256, 256)
	h.m.ioe = makeAndInitArray(256, ^byte(0))
	h.m.eob[0].v = offset
	h.m.eob[0].b = 0

	for i, c := range rank[1:] {
		if c == 0 {
			continue
		}
		offset += c
		h.m.eob[idx].v = offset
		h.m.eob[idx].b = byte(i + 1)
		h.m.ioe[i+1] = idx
		idx++
	}
	h.m.eob = h.m.eob[:idx]

	return h
}

// getBlockRange returns range (S, E] of char in BWT
// S, E -> are ZERO based offset marks the start and end offset of char range
func (m *meta) getBlockRange(b byte) (uint, uint, bool) {
	if b == 0 {
		// m.ioe[0] is 0
		return 0, m.eob[0].v - 1, true
	}
	i := m.ioe[b]
	if i == ^byte(0) {
		return 0, 0, false
	}
	return m.eob[i-1].v - 1, m.eob[i].v - 1, true
}

// ranks at previous blocks of character b,
// e -> offset of current block
func blockRank(b byte, e uint, s []super, bsz []uint16, char []byte, hist []uint16) uint {
	// get ranks from super block
	// i -> starting offset in the super block
	// r -> ranks of the super block
	i, r, j := e/sbsz, uint(0), uint(0)
	if i > 0 {
		i, r, j = i*sbsz, s[i-1].rank[b], s[i-1].offset
	}

	if e > i {
		sz := uint(0)
		for _, v := range bsz[i:e] {
			sz += uint(v)
		}

		k := j + sz - 1
		for ; k > j; k-- {
			r += uint(coefficient[char[k]^b]) * uint(hist[k])
		}
		r += uint(coefficient[char[k]^b]) * uint(hist[k])
	}

	return r
}

func makeAndInitArray(l int, v byte) []byte {
	a := make([]byte, l, l)
	a[0] = v
	for i := 1; i <= 128; i <<= 1 {
		copy(a[i:], a[:i])
	}

	return a
}

// note: same if freq equals, use alphabetical order, check CalcBlockHistogram for encoding
func findMostFreqChar(a []byte, hist []uint16) byte {
	i := 0
	for j := 1; j < len(hist); j++ {
		if hist[j] > hist[i] {
			i = j
		}
	}

	return a[i]
}
