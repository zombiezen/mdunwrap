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

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFilter(t *testing.T) {
	const testdataRoot = "testdata"
	entries, err := os.ReadDir(testdataRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, ent := range entries {
		const inSuffix = ".in.md"
		const outSuffix = ".out.md"
		name := ent.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, inSuffix) {
			continue
		}
		base := name[:len(name)-len(inSuffix)]

		t.Run(base, func(t *testing.T) {
			in, err := os.ReadFile(filepath.Join(testdataRoot, base+inSuffix))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(testdataRoot, base+outSuffix))
			if err != nil {
				t.Fatal(err)
			}
			got := filter(in)
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("-want +got:\n%s", diff)
			}
		})
	}
}
