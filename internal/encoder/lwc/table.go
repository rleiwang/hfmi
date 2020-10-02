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

var (
	single = make([][]byte, 256)
	half   = make([][]byte, 256)
	nibble = make([][]byte, 256)
)

func init() {
	buf := make([]byte, 256*8)
	for i := 0; i < 256; i++ {
		single[i], buf = buf[:8], buf[8:]
		single[i][0] = byte(i) & 0x01
		single[i][1] = byte(i>>1) & 0x01
		single[i][2] = byte(i>>2) & 0x01
		single[i][3] = byte(i>>3) & 0x01
		single[i][4] = byte(i>>4) & 0x01
		single[i][5] = byte(i>>5) & 0x01
		single[i][6] = byte(i>>6) & 0x01
		single[i][7] = byte(i>>7) & 0x01
	}

	buf = make([]byte, 256*4)
	for i := 0; i < 256; i++ {
		half[i], buf = buf[:4], buf[4:]
		half[i][0] = byte(i) & 0x03
		half[i][1] = byte(i>>2) & 0x03
		half[i][2] = byte(i>>4) & 0x03
		half[i][3] = byte(i>>6) & 0x03
	}

	buf = make([]byte, 256*2)
	for i := 0; i < 256; i++ {
		nibble[i], buf = buf[:2], buf[2:]
		nibble[i][0] = byte(i) & 0x0F
		nibble[i][1] = byte(i>>4) & 0x0F
	}
}
