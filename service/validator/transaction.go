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

package validator

import (
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps-rosetta/service/failure"
	"github.com/optakt/flow-dps-rosetta/service/identifier"
)

// Transaction validates a transaction identifier, and if its valid, returns a matching Flow Identifier.
func (v *Validator) Transaction(transaction identifier.Transaction) (flow.Identifier, error) {

	txID, err := flow.HexStringToIdentifier(transaction.Hash)
	if err != nil {
		return flow.ZeroID, failure.InvalidTransaction{
			Hash:        transaction.Hash,
			Description: failure.NewDescription(txHashInvalid),
		}
	}

	return txID, nil
}
