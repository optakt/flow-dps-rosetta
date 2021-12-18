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

	"github.com/onflow/flow-go/model/flow"
)

// InvalidProposer is the error for an invalid transaction proposer.
type InvalidProposer struct {
	Description Description
	Have        flow.Address
	Want        flow.Address
}

// Error implements the error interface.
func (i InvalidProposer) Error() string {
	return fmt.Sprintf("invalid transaction proposer (have: %s, want: %s): %s", i.Have.Hex(), i.Want.Hex(), i.Description)
}
