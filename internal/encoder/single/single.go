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

package single

import "github.com/rleiwang/hfmi/internal"

type single struct {
	c byte
	n uint16
}

func New(c byte, n uint16) internal.SDS {
	return &single{c, n}
}

func (s *single) Access(p uint, bv []byte) (byte, uint) {
	return s.c, p + 1
}

func (s *single) Rank(b byte, p uint, bv []byte) uint {
	if b == s.c {
		return p + 1
	}
	return 0
}
