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

// Data implements the Rosetta Data API specification.
// See https://www.rosetta-api.org/docs/data_api_introduction.html
type Data struct {
	config   Configuration
	retrieve Retriever
	validate Validator
}

// NewData creates a new instance of the Data API using the given configuration to answer configuration queries
// and the given retriever to answer blockchain data queries.
func NewData(config Configuration, retrieve Retriever, validate Validator) *Data {
	d := Data{
		config:   config,
		retrieve: retrieve,
		validate: validate,
	}
	return &d
}
