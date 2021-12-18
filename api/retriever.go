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

package api

import (
	"time"

	"github.com/optakt/flow-dps-rosetta/service/identifier"
	"github.com/optakt/flow-dps-rosetta/service/object"
)

type Retriever interface {
	Oldest() (identifier.Block, time.Time, error)
	Current() (identifier.Block, time.Time, error)
	Block(rosBlockID identifier.Block) (*object.Block, []identifier.Transaction, error)
	Transaction(rosBlockID identifier.Block, rosTxID identifier.Transaction) (*object.Transaction, error)
	Balances(rosBlockID identifier.Block, rosAccountID identifier.Account, rosCurrencies []identifier.Currency) (identifier.Block, []object.Amount, error)
	Sequence(rosBlockID identifier.Block, rosAccountID identifier.Account, index int) (uint64, error)
}
