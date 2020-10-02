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
	"testing"

	"github.com/rleiwang/hfmi/internal"
)

func TestAccess(t *testing.T) {
	type args struct {
		bv []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"textbook", args{[]byte("tobeornottobethatisthequestion")}},
	}
	r := &runlen{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chars, hist, mfc, runs := internal.CalcBlockHistogram(tt.args.bv)
			l := CompSZ(chars, hist, runs)
			bv := make([]byte, l)
			Encode(bv, tt.args.bv, mfc, chars, hist)

			ranks := [256]uint{}
			for i, c := range tt.args.bv {
				ranks[c]++
				gotc, gotr := r.Access(uint(i), bv)
				if gotc != c || gotr != ranks[c] {
					t.Errorf("runlen.Access() gotc = %v, gotr = %v, want c=%v, r=%v\n", gotc, gotr, c, ranks[c])
				}
			}
		})
	}
}

func TestRank(t *testing.T) {
	type args struct {
		bv []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"textbook", args{[]byte("tobeornottobethatisthequestion")}},
	}
	r := &runlen{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chars, hist, mfc, runs := internal.CalcBlockHistogram(tt.args.bv)
			l := CompSZ(chars, hist, runs)
			bv := make([]byte, l)
			Encode(bv, tt.args.bv, mfc, chars, hist)

			ranks := [256]uint{}
			for i, c := range tt.args.bv {
				ranks[c]++
				gotr := r.Rank(c, uint(i), bv)
				if gotr != ranks[c] {
					t.Errorf("runlen.Rank() gotr = %v, want r=%v\n", gotr, ranks[c])
				}
			}
		})
	}
}

func BenchmarkAccess(b *testing.B) {
	type args struct {
		bv []byte
	}
	benchmarks := []struct {
		name string
		args args
	}{
		{"textbook", args{[]byte("tobeornottobethatisthequestion")}},
	}
	for _, tt := range benchmarks {
		b.Run(tt.name, func(b *testing.B) {
			r := &runlen{}
			chars, hist, mfc, runs := internal.CalcBlockHistogram(tt.args.bv)
			l := CompSZ(chars, hist, runs)
			bv := make([]byte, l)
			Encode(bv, tt.args.bv, mfc, chars, hist)

			for i := 0; i < b.N; i++ {
				for i := 0; i < len(tt.args.bv); i++ {
					r.Access(uint(i), bv)
				}
			}
		})
	}
}

func BenchmarkRank(b *testing.B) {
	r := &runlen{}
	ba := []byte("tobeornottobethatisthequestion")
	chars, hist, mfc, runs := internal.CalcBlockHistogram(ba)
	l := CompSZ(chars, hist, runs)
	bv := make([]byte, l)
	Encode(bv, ba, mfc, chars, hist)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Rank('t', 20, bv)
	}
}
