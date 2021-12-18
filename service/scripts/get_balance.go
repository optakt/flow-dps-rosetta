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

package scripts

// Adopted from:
// https://github.com/onflow/flow-core-contracts/blob/master/transactions/flowToken/scripts/get_balance.cdc

const getBalance = `// This script reads the balance field of an account's FlowToken Balance

import FungibleToken from 0x{{.Params.FungibleToken}}
import {{.Token.Type}} from 0x{{.Token.Address}}

pub fun main(account: Address): UFix64 {

    let vaultRef = getAccount(account)
        .getCapability({{.Token.Balance}})
        .borrow<&{{.Token.Type}}.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
`
