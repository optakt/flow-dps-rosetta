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

//go:build integration
// +build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps/models/convert"
	"github.com/optakt/flow-dps/models/dps"

	rosetta "github.com/optakt/flow-dps-rosetta/api"
	"github.com/optakt/flow-dps-rosetta/service/configuration"
	"github.com/optakt/flow-dps-rosetta/service/identifier"
	"github.com/optakt/flow-dps-rosetta/service/object"
	"github.com/optakt/flow-dps-rosetta/service/request"
	"github.com/optakt/flow-dps-rosetta/service/response"
)

type validateBlockFunc func(identifier.Block)
type validateTxFunc func([]*object.Transaction)

func TestAPI_Block(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	// Headers of known blocks to verify.
	var (
		firstHeader  = knownHeader(0)
		secondHeader = knownHeader(1)
		midHeader1   = knownHeader(47)
		midHeader2   = knownHeader(60)
		midHeader3   = knownHeader(65)
		lastHeader   = knownHeader(173) // header of last indexed block
	)

	const (
		senderAccount         = "10c4fef62310c807"
		senderReceiverAccount = "e2f72218abeec2b9"
		receiverAccount       = "06909bc5ba14c266"

		firstTx  = "2d394a7841c91c5470e6e3cabb1e7ed57609ef41117bba84ced01d37659f2861"
		secondTx = "c059880060a66e84b23fbb8f2cd1fb24df64c9baad6e150ed8622e6eeb52031e"
	)

	tests := []struct {
		name string

		request request.Block

		wantTimestamp        int64
		wantParentHash       string
		wantParentHeight     uint64
		validateTransactions validateTxFunc
		validateBlock        validateBlockFunc
	}{
		{
			// First block. Besides the standard validation, it's also a special case
			// since according to the Rosetta spec, it should point to itself as the parent.
			name:    "first block",
			request: blockRequest(firstHeader),

			wantTimestamp:    convert.RosettaTime(firstHeader.Timestamp),
			wantParentHash:   firstHeader.ID().String(),
			wantParentHeight: firstHeader.Height,
			validateBlock:    validateByHeader(t, firstHeader),
		},
		{
			name:    "child of first block",
			request: blockRequest(secondHeader),

			wantTimestamp:    convert.RosettaTime(secondHeader.Timestamp),
			wantParentHash:   secondHeader.ParentID.String(),
			wantParentHeight: secondHeader.Height - 1,
			validateBlock:    validateByHeader(t, secondHeader),
		},
		{
			// First transfer of currency from that is not tied to an account creation.
			name:    "block mid-chain with transactions",
			request: blockRequest(midHeader1),

			wantTimestamp:        convert.RosettaTime(midHeader1.Timestamp),
			wantParentHash:       midHeader1.ParentID.String(),
			wantParentHeight:     midHeader1.Height - 1,
			validateBlock:        validateByHeader(t, midHeader1),
			validateTransactions: validateTransfer(t, firstTx, senderReceiverAccount, receiverAccount, 5_00000000),
		},
		{
			name:    "block mid-chain without transactions",
			request: blockRequest(midHeader2),

			wantTimestamp:    convert.RosettaTime(midHeader2.Timestamp),
			wantParentHash:   midHeader2.ParentID.String(),
			wantParentHeight: midHeader2.Height - 1,
			validateBlock:    validateByHeader(t, midHeader2),
		},
		{
			// Transaction between two users.
			name:    "second block mid-chain with transactions",
			request: blockRequest(midHeader3),

			wantTimestamp:        convert.RosettaTime(midHeader3.Timestamp),
			wantParentHash:       midHeader3.ParentID.String(),
			wantParentHeight:     midHeader3.Height - 1,
			validateBlock:        validateByHeader(t, midHeader3),
			validateTransactions: validateTransfer(t, secondTx, senderAccount, senderReceiverAccount, 5_00000000),
		},
		{
			name: "lookup of a block mid-chain by index only",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID:   identifier.Block{Index: &midHeader3.Height},
			},

			wantTimestamp:        convert.RosettaTime(midHeader3.Timestamp),
			wantParentHash:       midHeader3.ParentID.String(),
			wantParentHeight:     midHeader3.Height - 1,
			validateTransactions: validateTransfer(t, secondTx, senderAccount, senderReceiverAccount, 5_00000000),
			validateBlock:        validateBlock(t, midHeader3.Height, midHeader3.ID().String()), // verify that the returned block ID has both height and hash
		},
		{
			name:    "last indexed block",
			request: blockRequest(lastHeader),

			wantTimestamp:    convert.RosettaTime(lastHeader.Timestamp),
			wantParentHash:   lastHeader.ParentID.String(),
			wantParentHeight: lastHeader.Height - 1,
			validateBlock:    validateByHeader(t, lastHeader),
		},
		{
			name: "last indexed block by omitting block identifier",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID:   identifier.Block{},
			},

			wantTimestamp:    convert.RosettaTime(lastHeader.Timestamp),
			wantParentHash:   lastHeader.ParentID.String(),
			wantParentHeight: lastHeader.Height - 1,
			validateBlock:    validateByHeader(t, lastHeader),
		},
	}

	for _, test := range tests {

		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			rec, ctx, err := setupRecorder(blockEndpoint, test.request)
			require.NoError(t, err)

			err = data.Block(ctx)
			assert.NoError(t, err)

			var blockResponse response.Block
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &blockResponse))
			require.NotNil(t, blockResponse.Block)

			test.validateBlock(blockResponse.Block.ID)

			assert.Equal(t, test.wantTimestamp, blockResponse.Block.Timestamp)

			// Verify that the information about the parent block (index and hash) is correct.
			assert.Equal(t, test.wantParentHash, blockResponse.Block.ParentID.Hash)

			if test.validateTransactions != nil {
				test.validateTransactions(blockResponse.Block.Transactions)
			}

			require.NotNil(t, blockResponse.Block.ParentID.Index)
			assert.Equal(t, test.wantParentHeight, *blockResponse.Block.ParentID.Index)
		})
	}
}

func TestAPI_BlockHandlesErrors(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	var (
		validBlockHeight uint64 = 41
		lastHeight       uint64 = 173

		validBlockHash = knownHeader(validBlockHeight).ID().String()
	)

	const trimmedBlockHash = "f91704ce2fa9a1513500184ebfec884a1728438463c0104f8a17d5c66dd1af7" // blockID a character too short

	var validBlockID = identifier.Block{
		Index: &validBlockHeight,
		Hash:  validBlockHash,
	}

	tests := []struct {
		name string

		request request.Block

		checkErr assert.ErrorAssertionFunc
	}{
		{
			// Effectively the same as the 'missing blockchain name' test case, since it's the first validation step.
			name:    "empty block request",
			request: request.Block{},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "missing blockchain name",
			request: request.Block{
				NetworkID: identifier.Network{
					Blockchain: "",
					Network:    dps.FlowLocalnet.String(),
				},
				BlockID: validBlockID,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid blockchain name",
			request: request.Block{
				NetworkID: identifier.Network{
					Blockchain: invalidBlockchain,
					Network:    dps.FlowLocalnet.String(),
				},
				BlockID: validBlockID,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidNetwork),
		},
		{
			name: "missing network name",
			request: request.Block{
				NetworkID: identifier.Network{
					Blockchain: dps.FlowBlockchain,
					Network:    "",
				},
				BlockID: validBlockID,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid network name",
			request: request.Block{
				NetworkID: identifier.Network{
					Blockchain: dps.FlowBlockchain,
					Network:    invalidNetwork,
				},
				BlockID: validBlockID,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidNetwork),
		},
		{
			name: "invalid length of block id",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(43),
					Hash:  trimmedBlockHash,
				},
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid block hash",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(13),
					Hash:  invalidBlockHash,
				},
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidBlock),
		},
		{
			name: "unknown block",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(lastHeight + 1),
				},
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorUnknownBlock),
		},
		{
			name: "mismatched block height and hash",
			request: request.Block{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(validBlockHeight - 1),
					Hash:  validBlockHash,
				},
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidBlock),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			_, ctx, err := setupRecorder(blockEndpoint, test.request)
			require.NoError(t, err)

			err = data.Block(ctx)
			test.checkErr(t, err)
		})
	}
}

func TestAPI_BlockHandlesMalformedRequest(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	const (
		// Network field is an integer instead of a string.
		wrongFieldType = `
		{ 
			"network_identifier": { 
				"blockchain": "flow", 
				"network": 99
			}
		}`

		unclosedBracket = `
		{
			"network_identifier": {
				"blockchain": "flow",
				"network": "flow-localnet"
			},
			"block_identifier" : {
				"index" : 41,
				"hash" : "f91704ce2fa9a1513500184ebfec884a1728438463c0104f8a17d5c66dd1af79"
			}`

		validJSON = `
		{
			"network_identifier": {
				"blockchain": "flow",
				"network": "flow-localnet"
			},
			"block_identifier" : {
				"index" : 41,
				"hash" : "f91704ce2fa9a1513500184ebfec884a1728438463c0104f8a17d5c66dd1af79"
			}
		}`
	)

	tests := []struct {
		name    string
		payload []byte
		prepare func(*http.Request)
	}{
		{
			name:    "wrong field type",
			payload: []byte(wrongFieldType),
			prepare: func(req *http.Request) {
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			},
		},
		{
			name:    "unclosed bracket",
			payload: []byte(unclosedBracket),
			prepare: func(req *http.Request) {
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			},
		},
		{
			name:    "valid payload with no MIME type set",
			payload: []byte(validJSON),
			prepare: func(req *http.Request) {
				req.Header.Set(echo.HeaderContentType, "")
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			_, ctx, err := setupRecorder(blockEndpoint, test.payload, test.prepare)
			require.NoError(t, err)

			err = data.Block(ctx)
			assert.Error(t, err)

			echoErr, ok := err.(*echo.HTTPError)
			require.True(t, ok)

			assert.Equal(t, http.StatusBadRequest, echoErr.Code)

			gotErr, ok := echoErr.Message.(rosetta.Error)
			require.True(t, ok)

			assert.Equal(t, configuration.ErrorInvalidEncoding, gotErr.ErrorDefinition)
			assert.NotEmpty(t, gotErr.Description)
		})
	}

}

// blockRequest generates a BlockRequest with the specified parameters.
func blockRequest(header flow.Header) request.Block {

	return request.Block{
		NetworkID: defaultNetwork(),
		BlockID: identifier.Block{
			Index: &header.Height,
			Hash:  header.ID().String(),
		},
	}
}

func validateTransfer(t *testing.T, hash string, from string, to string, amount int64) validateTxFunc {

	t.Helper()

	return func(transactions []*object.Transaction) {

		require.Len(t, transactions, 1)

		tx := transactions[0]

		assert.Equal(t, tx.ID.Hash, hash)

		// Operations come in pairs. A negative transfer of funds for the sender and a positive one for the receiver.
		require.Equal(t, len(tx.Operations), 2)

		op1 := tx.Operations[0]
		op2 := tx.Operations[1]

		assert.Equal(t, op1.Type, dps.OperationTransfer)
		assert.Equal(t, op1.Status, dps.StatusCompleted)

		assert.Equal(t, op1.Amount.Currency.Symbol, dps.FlowSymbol)
		assert.Equal(t, op1.Amount.Currency.Decimals, uint(dps.FlowDecimals))

		address := op1.AccountID.Address
		if address != from && address != to {
			t.Errorf("unexpected account address (%v)", address)
		}

		wantValue := strconv.FormatInt(amount, 10)
		if address == from {
			wantValue = "-" + wantValue
		}

		assert.Equal(t, op1.Amount.Value, wantValue)

		assert.Equal(t, op2.Type, dps.OperationTransfer)
		assert.Equal(t, op2.Status, dps.StatusCompleted)

		assert.Equal(t, op2.Amount.Currency.Symbol, dps.FlowSymbol)
		assert.Equal(t, op2.Amount.Currency.Decimals, uint(dps.FlowDecimals))

		address = op2.AccountID.Address
		if address != from && address != to {
			t.Errorf("unexpected account address (%v)", address)
		}

		wantValue = strconv.FormatInt(amount, 10)
		if address == from {
			wantValue = "-" + wantValue
		}

		assert.Equal(t, op2.Amount.Value, wantValue)
	}
}
