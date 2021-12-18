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

package response

import (
	"github.com/optakt/flow-dps-rosetta/service/identifier"
	"github.com/optakt/flow-dps-rosetta/service/object"
)

// Balance implements the successful response schema for /account/balance.
// See https://www.rosetta-api.org/docs/AccountApi.html#200---ok
type Balance struct {
	BlockID  identifier.Block `json:"block_identifier"`
	Balances []object.Amount  `json:"balances"`
}
