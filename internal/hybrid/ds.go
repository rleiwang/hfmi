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

import "github.com/rleiwang/hfmi/internal"

// encoding type
type edt byte

const (
	single edt = iota
	runlen
	sparse
	lwc
)

const (
	sbsz = 8            // super block size, number of blocks
	msb  = byte(1) << 7 // most significant bit
)

const (
	αsz = 256
)

type pair struct {
	v uint
	b byte
}

type super struct {
	rank   [256]uint
	offset uint
}

type meta struct {
	eob   []pair // end of bucket
	ioe   []byte // index of end of buckets
	bsds  []internal.SDS
	bsz   []uint16
	bbv   [][]byte
	char  []byte
	hist  []uint16
	super []super // super header, absolute rank
}

type dictionary struct {
	fidx []byte // forward index: byte -> ith char
	ridx []byte // reverse index: ith char -> byte
}

type hybrid struct {
	cnt  uint        // total count
	hdr  []byte      // compressed header
	bv   []byte      // compressed bit vector
	dict *dictionary // dictionary
	m    meta        //
}
