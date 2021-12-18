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

package failure

import (
	"fmt"
)

// InvalidBlockHash is the error for a block hash of invalid length.
type InvalidBlockHash struct {
	Description Description
	WantLength  int
	HaveLength  int
}

// Error implements the error interface.
func (i InvalidBlockHash) Error() string {
	return fmt.Sprintf("invalid block hash length (want: %d, have: %d): %s", i.WantLength, i.HaveLength, i.Description)
}
