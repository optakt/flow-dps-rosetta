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

package retriever

import (
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go/model/flow"
)

// Invoker represents something that can retrieve public keys at any given height, and
// execute scripts to retrieve values from the Flow Virtual Machine.
type Invoker interface {
	Key(height uint64, address flow.Address, index int) (*flow.AccountPublicKey, error)
	Script(height uint64, script []byte, parameters []cadence.Value) (cadence.Value, error)
}
