// Copyright 2021 Optakt Labs OÜ
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package convert_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/optakt/flow-dps-rosetta/testing/mocks"
	"github.com/optakt/flow-dps/models/convert"
)

func TestTypesToStrings(t *testing.T) {
	types := mocks.GenericEventTypes(4)

	got := convert.TypesToStrings(types)

	for _, typ := range types {
		assert.Contains(t, got, string(typ))
	}
}

func TestStringsToTypes(t *testing.T) {
	types := mocks.GenericEventTypes(4)

	var ss []string
	for _, typ := range types {
		ss = append(ss, string(typ))
	}

	got := convert.StringsToTypes(ss)

	assert.Equal(t, types, got)
}

func TestRosettaTime(t *testing.T) {
	ti := time.Now()

	got := convert.RosettaTime(ti)

	assert.Equal(t, ti.UnixNano()/1_000_000, got)
}
