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

package request

import (
	"github.com/optakt/flow-dps-rosetta/service/identifier"
)

// Parse implements the request schema for /construction/parse.
// See https://www.rosetta-api.org/docs/ConstructionApi.html#request-4
type Parse struct {
	NetworkID   identifier.Network `json:"network_identifier"`
	Signed      bool               `json:"signed"`
	Transaction string             `json:"transaction"`
}
