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

package object

import (
	"github.com/optakt/flow-dps-rosetta/service/identifier"
)

// Options object is used in the Rosetta Construction API requests.
// This object is returned in the `/construction/preprocess` response,
// and is forwarded to the `/construction/metadata` endpoint unmodified.
//
// Specifically for Flow DPS, this object contains the account identifier
// that is the proposer of the transaction (by default, this is the sender).
// Account identifier is required so that we can return the sequence number
// of the proposer's key, required for the Flow transaction.
type Options struct {
	AccountID identifier.Account `json:"account_identifier"`
}
