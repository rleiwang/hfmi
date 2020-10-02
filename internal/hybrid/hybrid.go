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
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"sort"

	"github.com/rleiwang/hfmi/internal"
)

func (h *hybrid) Access(p uint) (a byte, r uint, ok bool) {
	if p < h.cnt {
		b, r := h.m.bsds[p/internal.SZ].Access(p%internal.SZ, h.m.bbv[p/internal.SZ])
		return h.dict.ridx[b], r + blockRank(b, p/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist), true
	}
	return 0, 0, false
}

func (h *hybrid) Locate(p uint) (byte, uint, bool) {
	if p >= h.cnt {
		return 0, 0, false
	}

	// note: eob's v is exclusive v), locate the bucket index
	i := sort.Search(len(h.m.eob), func(i int) bool { return h.m.eob[i].v > p })
	if i == 0 {
		return 0, p + 1, true
	}

	// return the bucket char, ranks = offset from begining of the bucket
	return h.dict.ridx[h.m.eob[i].b], p - h.m.eob[i-1].v + 1, true
}

func (h *hybrid) Select(a byte, r uint) (p uint, ok bool) {
	// leave it blank
	return 0, false
}

func (h *hybrid) Rank(a byte, p uint) (r uint, ok bool) {
	b := h.dict.fidx[a]
	r = h.m.bsds[p/internal.SZ].Rank(b, p%internal.SZ, h.m.bbv[p/internal.SZ])
	return blockRank(b, p/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist) + r, true
}

func (h *hybrid) Dictionary() []byte {
	return h.dict.ridx
}

func (h *hybrid) Bytes() []byte {
	b := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(b, uint32(len(h.hdr)))

	b = append(b, h.hdr...)
	return append(b, h.bv...)
}

func (h *hybrid) Count(p string) uint {
	rng, ok := h.Search(p)
	if !ok {
		return 0
	}

	// (s, e]
	return rng[1] - rng[0]
}

func (h *hybrid) Search(p string) ([]uint, bool) {
	pat := []byte(p)
	if len(pat) == 0 {
		return nil, false
	}
	for i, p := range pat {
		pat[i] = h.dict.fidx[p]
	}

	// range is (s, e]
	s, e, ok := h.m.getBlockRange(pat[0])
	if !ok {
		return nil, ok
	}
	for _, b := range pat[1:] {
		if s == e {
			break
		}
		offset, _, ok := h.m.getBlockRange(b)
		if !ok {
			return nil, ok
		}
		s = offset + blockRank(b, s/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist) +
			h.m.bsds[s/internal.SZ].Rank(b, s%internal.SZ, h.m.bbv[s/internal.SZ])
		e = offset + blockRank(b, e/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist) +
			h.m.bsds[e/internal.SZ].Rank(b, e%internal.SZ, h.m.bbv[e/internal.SZ])
	}

	if e > s {
		return []uint{s, e}, true
	}

	return nil, false
}

func (h *hybrid) Size() (int, int) {
	return len(h.hdr), len(h.bv)
}

func (h *hybrid) Len() uint {
	return h.cnt
}

func (h *hybrid) Histogram() []uint {
	chars := make([]uint, αsz, αsz)
	for _, c := range h.m.char {
		chars[h.dict.ridx[c]] += uint(h.m.hist[c])
	}

	return chars
}

func (h *hybrid) CharsInBound(s, e uint) []byte {
	s /= internal.SZ
	e /= internal.SZ

	// i -> starting offset in the super block
	i, offset := s/sbsz, uint(0)
	if i > 0 {
		i, offset = i*sbsz, h.m.super[i-1].offset
	}

	for _, v := range h.m.bsz[i:s] {
		offset += uint(v)
	}

	sz := uint(0)
	for _, v := range h.m.bsz[s : e+1] {
		sz += uint(v)
	}

	chars := [256]byte{}
	for _, c := range h.m.char[offset : offset+sz] {
		chars[c] = 1
	}

	offset = 0
	for i, c := range chars {
		if c > 0 {
			chars[offset] = h.dict.ridx[i]
			offset++
		}
	}
	return chars[:offset]
}

func (h *hybrid) GetBound(b byte) (uint, uint, bool) {
	nb := h.dict.fidx[b]
	if nb == 255 && b != 255 {
		return 0, 0, false
	}

	return h.m.getBlockRange(nb)
}

func (h *hybrid) Restore(w io.Writer) bool {
	pgSZ := os.Getpagesize()
	nxtZero, nxtOne, _ := h.m.getBlockRange(0)
	nxtOne++
	np, eod, buf, j := nxtZero, nxtOne, make([]byte, pgSZ, pgSZ), 0

loop:
	for {
		b, r := h.m.bsds[np/internal.SZ].Access(np%internal.SZ, h.m.bbv[np/internal.SZ])

		switch b {
		case 0:
			nxtZero++
			if nxtZero == eod {
				break loop
			}
			np = nxtZero
		case 1:
			// for separators, just go straight to the next char
			np = nxtOne
			nxtOne++
			buf[j] = ' '
		default:
			buf[j] = h.dict.ridx[b]
			offset, _, _ := h.m.getBlockRange(b)
			np = offset + r + blockRank(b, np/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist)
		}

		j = (j + 1) % pgSZ
		if j == 0 {
			w.Write(buf)
		}
	}

	if j > 0 {
		w.Write(buf[:j])
	}

	return true
}

func (h *hybrid) ForwardExtractToChar(p uint, t byte) ([]byte, uint, bool) {
	nt := h.dict.fidx[t]
	if nt == 255 && t != 255 {
		// t doesn't exist, to the end
		nt = 0
	}

	buf := bytes.NewBuffer(make([]byte, 0, h.cnt))
	for {
		b, r := h.m.bsds[p/internal.SZ].Access(p%internal.SZ, h.m.bbv[p/internal.SZ])

		if b == nt || b == 0 {
			return buf.Bytes(), p, true
		}

		offset, _, _ := h.m.getBlockRange(b)
		p = r + blockRank(b, p/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist) + offset

		buf.WriteByte(h.dict.ridx[b])
	}
	return nil, 0, false
}

func (h *hybrid) BackwardExtractToChar(p uint, t byte) ([]byte, uint, bool) {
	var buf bytes.Buffer
	for {
		b, r, ok := h.Locate(p)
		if !ok {
			return nil, 0, false
		}
		if b == t {
			break
		} else if b == 0 {
			p = 0
			break
		}
		buf.WriteByte(b)
		if p, ok = h.Select(b, r); !ok {
			return nil, 0, false
		}
	}

	d := buf.Bytes()
	s, e := 0, len(d)-1
	for s < e {
		d[s], d[e] = d[e], d[s]
		s++
		e--
	}

	return d, p, true
}

func (h *hybrid) BackwardJumpToChar(p uint, t byte) (uint, bool) {
	for {
		b, r, ok := h.Locate(p)
		if !ok {
			return 0, false
		}
		if b == t {
			return p, true
		} else if b == 0 {
			return 0, true
		}
		if p, ok = h.Select(b, r); !ok {
			return 0, false
		}
	}
}

func (h *hybrid) ExtractFields(sep byte, p uint, fc uint) ([][]byte, bool) {
	// forward
	fbuf, _, ok := h.ForwardExtractToChar(p, sep)
	if !ok {
		return nil, false
	}

	var buf []byte
	b, _, ok := h.Locate(p)
	if !ok {
		return nil, false
	}
	if b > sep {
		// p is not in sep bucket, before BackwardExtractToChar
		buf, p, ok = h.BackwardExtractToChar(p, sep)
		// p is in sep bucket after BackwardExtractToChar
		if !ok {
			return nil, false
		}
	}

	// p is in sep bucket
	offset, _, ok := h.m.getBlockRange(sep)
	if !ok {
		return nil, false
	}

	// r is sep ranks
	r := p - offset
	ret := make([][]byte, fc, fc)
	ret[r%fc] = append(buf, fbuf...)

	// ith line
	ith := r / fc
	for s, e := ith*fc, (ith+1)*fc; s < e; s++ {
		if s == r {
			continue
		}
		buf, _, ok = h.ForwardExtractToChar(offset+s, sep)
		if !ok {
			return nil, false
		}
		ret[s%fc] = buf
	}

	return ret, true
}

func (h *hybrid) ExtractAllFields(sep byte, fc uint) ([][][]byte, bool) {
	// get number of fields
	beg, end, ok := h.m.getBlockRange(sep)
	if !ok {
		return nil, false
	}

	// get number of rows
	ret := make([][][]byte, (end-beg)/fc, (end-beg)/fc)
	for ith := range ret {
		ret[ith] = make([][]byte, fc, fc)
		for j := uint(0); j < fc; j++ {
			p := uint(ith)*fc + j
			if buf, _, ok := h.ForwardExtractToChar(p, sep); ok {
				ret[ith][j] = buf
			} else {
				ret[ith][j] = []byte{}
			}
		}
	}

	return ret, true
}

func (h *hybrid) ExtractRange(from, to uint) ([]byte, bool) {
	var buf bytes.Buffer
	for {
		b, r := h.m.bsds[from/internal.SZ].Access(from%internal.SZ, h.m.bbv[from/internal.SZ])
		if b < 2 {
			// reached byte 0 or 1
			return buf.Bytes(), true
		}

		buf.WriteByte(h.dict.ridx[b])

		if from == to {
			return buf.Bytes(), true
		}

		offset, _, _ := h.m.getBlockRange(b)
		from = offset + r + blockRank(b, from/internal.SZ, h.m.super, h.m.bsz, h.m.char, h.m.hist)
	}
	return nil, false
}
