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

	"github.com/rleiwang/sa"

	"github.com/rleiwang/hfmi"
	"github.com/rleiwang/hfmi/internal"
	lwcenc "github.com/rleiwang/hfmi/internal/encoder/lwc"
	rlenc "github.com/rleiwang/hfmi/internal/encoder/runlen"
	spenc "github.com/rleiwang/hfmi/internal/encoder/sparse"
)

// New build FMI index from text t
func New(t []byte) hfmi.FMI {
	// eob, end of bwt
	_, bwt, aux := sa.BWT(t)
	// aux.Dict -> reverse index, fidx -> forward index
	fidx := newForwardIndex(aux.Dict)

	return buildFMI(bwt, &dictionary{fidx: fidx, ridx: aux.Dict})
}

func Build(cnt uint, ridx, d []byte) hfmi.FMI {
	offset := binary.LittleEndian.Uint32(d[:4]) + 4
	return restoreHeader(&hybrid{
		cnt:  cnt,
		hdr:  d[4:offset],
		bv:   d[offset:],
		dict: &dictionary{fidx: newForwardIndex(ridx), ridx: ridx},
	})
}

func newForwardIndex(ridx []byte) []byte {
	fidx := make([]byte, 256, 256)
	fidx[0] = 255
	for i := 1; i <= 128; i <<= 1 {
		copy(fidx[i:], fidx[:i])
	}

	for i, b := range ridx {
		fidx[b] = byte(i)
	}

	return fidx
}

func buildFMI(bwt []byte, dict *dictionary) hfmi.FMI {
	for i, b := range bwt {
		bwt[i] = dict.fidx[b]
	}

	blocks := split(bwt, internal.SZ)

	// max header sz: (3 + 256 * 2) * len(blocks), max bv sz: len(bwt)
	buf := make([]byte, 2*internal.SZ*len(blocks)+4)
	header, bv := buf[:4+len(buf)/2], buf[4+len(buf)/2:]
	hdrBeg, bvBeg, count := uint(4), uint(0), uint32(0)
	for _, b := range blocks {
		chars, hist, mfc, runs := internal.CalcBlockHistogram(b)
		count += uint32(len(chars))
		e, s := single, uint(0)
		var enc internal.Encoder
		if runs > 1 {
			e, enc = minSZ(chars, hist, runs)
			s = enc(bv[bvBeg:], b, mfc, chars, hist)
			bvBeg += s
		}
		hsz := encodeHeader(chars, hist, e, s, header[hdrBeg:])
		hdrBeg += hsz
	}

	binary.LittleEndian.PutUint32(buf, count)
	return restoreHeader(&hybrid{
		cnt:  uint(len(bwt)),
		hdr:  header[:hdrBeg],
		bv:   bv[:bvBeg],
		dict: dict,
	})
}

// a -> [char]=freq, runs -> number of runs
// return
// edt -> encoding type (single, runlen length, lwc or sparse)
// s -> size in byte in bv
func minSZ(chars []byte, hist []uint16, runs uint) (edt, internal.Encoder) {
	e, sz, enc := runlen, rlenc.CompSZ(chars, hist, runs), rlenc.Encode
	if ssz := spenc.CompSZ(chars, hist, runs); sz > ssz {
		e, sz, enc = sparse, ssz, spenc.Encode
	}

	if sz <= 16 {
		return e, enc
	}

	return lwc, lwcenc.Encode
}

func split(data []byte, sz int) [][]byte {
	cnt := (len(data) + sz - 1) / sz
	chunks := make([][]byte, cnt, cnt)
	for i := range chunks[:cnt-1] {
		chunks[i], data = data[:sz], data[sz:]
	}
	chunks[cnt-1] = data
	return chunks
}
