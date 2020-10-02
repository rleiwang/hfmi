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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/rleiwang/hfmi/internal"
)

var (
	chars [256]byte
)

func TestMain(m *testing.M) {
	for i := range chars {
		chars[i] = byte(255 - i)
	}
	internal.InitSegmentCache(1)
	os.Exit(m.Run())
}

func TestEncodeExpandFull(t *testing.T) {
	for i := 1; i <= 256; i++ {
		t.Run(fmt.Sprintf("full freq %d", i), func(t *testing.T) {
			want := make([]byte, 256)
			chars, hist := prepare(i, want)
			dst := make([]byte, CompSZ(chars, hist, 0))
			Encode(dst, want, 0, chars, hist)

			if got := Expand(dst, chars); !reflect.DeepEqual(got, want) {
				t.Errorf("Expand() = %v, want %v", got, want)
			}
		})
	}
}

func TestEncodeExpandPartial(t *testing.T) {
	for f := 1; f <= 256; f++ {
		for partial := 1; partial <= 256; partial++ {
			t.Run(fmt.Sprintf("freq %d partial %v", f, partial),
				func(t *testing.T) {
					want := make([]byte, 256)
					freq := f
					if partial < freq {
						freq = partial
					}
					chars, hist := prepare(freq, want[:partial])
					dst := make([]byte, CompSZ(chars, hist, 0))
					s := Encode(dst, want[:partial], 0, chars, hist)

					if got := Expand(dst[:s], chars); !reflect.DeepEqual(got[:partial], want[:partial]) {
						t.Errorf("Expand() = %v,\n want %v", got, want)
					}
				})
		}
	}
}

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
	r := &lwc{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chars, hist, mfc, _ := internal.CalcBlockHistogram(tt.args.bv)
			bv := make([]byte, 256)
			s := Encode(bv, tt.args.bv, mfc, chars, hist)

			expanded := Expand(bv[:s], chars)

			ranks := [256]uint{}
			for i, c := range tt.args.bv {
				ranks[c]++
				gotc, gotr := r.Access(uint(i), expanded)
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
	r := &lwc{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chars, hist, mfc, _ := internal.CalcBlockHistogram(tt.args.bv)
			bv := make([]byte, 256)
			s := Encode(bv, tt.args.bv, mfc, chars, hist)
			expanded := Expand(bv[:s], chars)

			ranks := [256]uint{}
			for i, c := range tt.args.bv {
				ranks[c]++
				gotr := r.Rank(c, uint(i), expanded)
				if gotr != ranks[c] {
					t.Errorf("runlen.Rank() gotr = %v, want r=%v\n", gotr, ranks[c])
				}
			}
		})
	}
}

func BenchmarkRank(b *testing.B) {
	r := &lwc{}
	ba := []byte("tobeornottobethatisthequestion")
	chars, hist, mfc, runs := internal.CalcBlockHistogram(ba)
	l := CompSZ(chars, hist, runs)
	bv := make([]byte, l)
	s := Encode(bv, ba, mfc, chars, hist)
	expanded := Expand(bv[:s], chars)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Rank('t', 20, expanded)
	}
}

func prepare(freq int, data []byte) ([]byte, []uint16) {
	chars, hist := make([]byte, freq), make([]uint16, freq)
	for i := 0; i < freq; i++ {
		chars[i] = chars[i]
	}

	for len(data) >= freq {
		for i := 0; i < freq; i++ {
			data[i] = chars[i]
		}
		data = data[freq:]
	}
	for i := 0; i < len(data); i++ {
		data[i] = chars[i]
	}

	return chars, hist
}
