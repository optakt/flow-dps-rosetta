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

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/optakt/flow-dps/models/dps"
)

// Generator dynamically generates Cadence scripts from templates.
type Generator struct {
	params          dps.Params
	getBalance      *template.Template
	transferTokens  *template.Template
	tokensDeposited *template.Template
	tokensWithdrawn *template.Template
}

// NewGenerator returns a Generator using the given parameters.
func NewGenerator(params dps.Params) *Generator {
	g := Generator{
		params:          params,
		getBalance:      template.Must(template.New("get_balance").Parse(getBalance)),
		transferTokens:  template.Must(template.New("transfer_tokens").Parse(transferTokens)),
		tokensDeposited: template.Must(template.New("tokensDeposited").Parse(tokensDeposited)),
		tokensWithdrawn: template.Must(template.New("withdrawal").Parse(tokensWithdrawn)),
	}
	return &g
}

// GetBalance generates a Cadence script to retrieve the balance of an account.
func (g *Generator) GetBalance(symbol string) ([]byte, error) {
	return g.bytes(g.getBalance, symbol)
}

// TransferTokens generates a Cadence script to operate a token transfer transaction.
func (g *Generator) TransferTokens(symbol string) ([]byte, error) {
	return g.bytes(g.transferTokens, symbol)
}

// TokensDeposited generates a Cadence script that matches the Flow event for tokens being deposited.
func (g *Generator) TokensDeposited(symbol string) (string, error) {
	return g.string(g.tokensDeposited, symbol)
}

// TokensWithdrawn generates a Cadence script that matches the Flow event for tokens being withdrawn.
func (g *Generator) TokensWithdrawn(symbol string) (string, error) {
	return g.string(g.tokensWithdrawn, symbol)
}

func (g *Generator) string(template *template.Template, symbol string) (string, error) {
	buf, err := g.compile(template, symbol)
	if err != nil {
		return "", fmt.Errorf("could not compile template: %w", err)
	}
	return buf.String(), nil
}

func (g *Generator) bytes(template *template.Template, symbol string) ([]byte, error) {
	buf, err := g.compile(template, symbol)
	if err != nil {
		return nil, fmt.Errorf("could not compile template: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *Generator) compile(template *template.Template, symbol string) (*bytes.Buffer, error) {
	token, ok := g.params.Tokens[symbol]
	if !ok {
		return nil, fmt.Errorf("invalid token symbol (%s)", symbol)
	}
	data := struct {
		Params dps.Params
		Token  dps.Token
	}{
		Params: g.params,
		Token:  token,
	}
	buf := &bytes.Buffer{}
	err := template.Execute(buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not execute template: %w", err)
	}
	return buf, nil
}
