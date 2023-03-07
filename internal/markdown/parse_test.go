// Copyright 2023 Ross Light
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		 https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package markdown

import "testing"

func TestIsThematicBreak(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"", -1},
		{"---\n", 3},
		{"***\n", 3},
		{"___\n", 3},
		{"+++\n", -1},
		{"===\n", -1},
		{"--\n", -1},
		{"**\n", -1},
		{"__\n", -1},
		{"_____________________________________\n", 37},
		{"- - -\n", 5},
		{"**  * ** * ** * **\n", 18},
		{"-     -      -      -\n", 21},
		{"- - - -    \n", 7},
		{"_ _ _ _ a\n", -1},
		{"a------\n", -1},
		{"---a---\n", -1},
		{"*-*\n", -1},
	}
	for _, test := range tests {
		if got := parseThematicBreak([]byte(test.line)); got != test.want {
			t.Errorf("isThematicBreak(%q) = %d; want %d", test.line, got, test.want)
		}
	}
}
