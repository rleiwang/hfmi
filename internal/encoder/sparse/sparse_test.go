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
    "fmt"
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
        {"textbook example", args{[]byte("tobeornottobethatisthequestion")}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            chars, hist, mfc, _ := internal.CalcBlockHistogram(tt.args.bv)
            dst := make([]byte, 256)
            r := Encode(dst, tt.args.bv, mfc, chars, hist)
            s := &sparse{mfc}

            ranks := [256]uint{}
            for i, c := range tt.args.bv {
                ranks[c]++
                gotc, gotr := s.Access(uint(i), dst[:r])
                if gotc != c || gotr != ranks[c] {
                    t.Errorf("sparse.Access() gotc = %v, gotr = %v, want c=%v, r=%v\n", gotc, gotr, c, ranks[c])
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
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            chars, hist, mfc, _ := internal.CalcBlockHistogram(tt.args.bv)
            dst := make([]byte, 256)
            r := Encode(dst, tt.args.bv, mfc, chars, hist)
            s := &sparse{mfc}

            ranks := [256]uint{}
            for i, c := range tt.args.bv {
                ranks[c]++
                gotr := s.Rank(c, uint(i), dst[:r])
                if gotr != ranks[c] {
                    t.Errorf("sparse.Rank() gotr = %v, want r=%v\n", gotr, ranks[c])
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
        fmt.Printf("%v\n", len(tt.args.bv))
        chars, hist, mfc, _ := internal.CalcBlockHistogram(tt.args.bv)
        dst := make([]byte, 256)
        r := Encode(dst, tt.args.bv, mfc, chars, hist)
        s := &sparse{mfc}
        b.StartTimer()

        b.Run(tt.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                for i := 0; i < len(tt.args.bv); i++ {
                    s.Access(uint(i), dst[:r])
                }
            }
        })
    }
}
