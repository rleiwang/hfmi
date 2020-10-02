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
	"io/ioutil"
	"path"
	"testing"

	"github.com/rleiwang/sa"
)

func TestAccess(t *testing.T) {
	type args struct {
		t []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"textbook", args{[]byte("tobeornottobethatisthequestion")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := make([]byte, len(tt.args.t))
			copy(orig, tt.args.t)
			_, bwt, _ := sa.BWT(orig)
			fmi := New(tt.args.t)
			buf := fmi.Bytes()
			fmi = Build(uint(len(bwt)), fmi.Dictionary(), buf)
			ranks := [256]uint{}
			for i, c := range bwt {
				ranks[c]++
				if b, r, ok := fmi.Access(uint(i)); !ok || b != c || r != ranks[c] {
					t.Errorf("Access(%v) = %v, %v, %v, want %v, %v", i, b, r, ok, c, ranks[c])
				}
			}
		})
	}
}

func TestRank(t *testing.T) {
	type args struct {
		t []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{"two", args{[]byte{32, 16}}},
		{"one", args{[]byte{32}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := make([]byte, len(tt.args.t))
			copy(orig, tt.args.t)
			_, bwt, _ := sa.BWT(orig)
			fmi := New(tt.args.t)
			buf := fmi.Bytes()
			fmi = Build(uint(len(bwt)), fmi.Dictionary(), buf)
			ranks := [256]uint{}
			for i, c := range bwt {
				ranks[c]++
				if r, ok := fmi.Rank(c, uint(i)); !ok || r != ranks[c] {
					t.Errorf("Rank(%v) = %v, %v, want %v, %v", i, r, ok, c, ranks[c])
				}
			}
		})
	}
}

func BenchmarkRank(b *testing.B) {
	index := New([]byte("tobeornottobethatisthequestion"))
	restoreHeader(index.(*hybrid))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.Rank('e', 15)
	}
}

func readTestFile(file string) []byte {
	content, err := ioutil.ReadFile(path.Join("../testdata", file))
	if err != nil {
		panic(err)
	}
	return content
}
