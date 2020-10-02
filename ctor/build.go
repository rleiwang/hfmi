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

package ctor

import (
	"github.com/rleiwang/hfmi"
	"github.com/rleiwang/hfmi/internal"
	"github.com/rleiwang/hfmi/internal/hybrid"
)

// New construct FM-Index from BWT
func New(t []byte) hfmi.FMI {
	return hybrid.New(t)
}

// Build restore serialized FM-index from bytes
func Build(cnt uint, ridx, d []byte) hfmi.FMI {
	return hybrid.Build(cnt, ridx, d)
}

// SetSegmentCache must set to a value corresponds to the number of processing goroutines
func SetSegmentCache(sz uint) {
	internal.InitSegmentCache(uint64(sz))
}
