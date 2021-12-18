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
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/flow-go/model/flow"

	rosetta "github.com/optakt/flow-dps-rosetta/api"
	"github.com/optakt/flow-dps-rosetta/service/configuration"
	"github.com/optakt/flow-dps-rosetta/service/identifier"
	"github.com/optakt/flow-dps-rosetta/service/object"
	"github.com/optakt/flow-dps-rosetta/service/request"
	"github.com/optakt/flow-dps-rosetta/service/response"
	"github.com/optakt/flow-dps-rosetta/testing/mocks"
	"github.com/optakt/flow-dps/models/dps"
)

func TestAPI_Transaction(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	var (
		firstHeader      = knownHeader(47)
		multipleTxHeader = knownHeader(57)
		lastHeader       = knownHeader(164)

		// two transactions in a single block
		midBlockTxs = []string{
			"9d8fb8f7d55ee3904fe5dc846236bafeac50fc64eb528d5e0eb1f693bdfd47d4",
			"9cb22148c60e23001dc1d22a8d16fa74bb6363674e2b1a8f6f1c02b34a9a5e11",
		}
	)

	const (
		firstTx = "2d394a7841c91c5470e6e3cabb1e7ed57609ef41117bba84ced01d37659f2861"
		lastTx  = "d7b8696b9a73550c228168d1fc5b771d35356d10eb7bba98edd1408d36a2f92b"
	)

	tests := []struct {
		name string

		request    request.Transaction
		validateTx validateTxFunc
	}{
		{
			name:       "some cherry picked transaction",
			request:    requestTransaction(firstHeader, firstTx),
			validateTx: validateTransfer(t, firstTx, "e2f72218abeec2b9", "06909bc5ba14c266", 5_00000000),
		},
		{
			// The test does not have blocks with more than two transactions, so this is the same as 'get the last transaction from a block'.
			name:       "second in a block with multiple",
			request:    requestTransaction(multipleTxHeader, midBlockTxs[1]),
			validateTx: validateTransfer(t, midBlockTxs[1], "06909bc5ba14c266", "10c4fef62310c807", 5_00000000),
		},
		{
			name:       "last transaction recorded",
			request:    requestTransaction(lastHeader, lastTx),
			validateTx: validateTransfer(t, lastTx, "1beecc6fef95b62e", "10c4fef62310c807", 5_00000000),
		},
	}

	for _, test := range tests {

		test := test
		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			rec, ctx, err := setupRecorder(transactionEndpoint, test.request)
			require.NoError(t, err)

			err = data.Transaction(ctx)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, rec.Result().StatusCode)

			var res response.Transaction
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

			test.validateTx([]*object.Transaction{res.Transaction})
		})
	}
}

func TestAPI_TransactionHandlesErrors(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	var testHeight uint64 = 105

	const (
		lastHeight = 173

		testBlockHash = "344368ba77ba47fb0a062dc8610c46cb2fe1539dbdedb0ba5fe8b46c629b0628"
		testTxHash    = "88419614bf6cda15586bb686f33eea15835db13c0f9f997dcce275afb325102a"

		trimmedBlockHash = "344368ba77ba47fb0a062dc8610c46cb2fe1539dbdedb0ba5fe8b46c629b062"  // block hash a character short
		trimmedTxHash    = "88419614bf6cda15586bb686f33eea15835db13c0f9f997dcce275afb325102"  // tx hash a character short
		invalidTxHash    = "88419614bf6cda15586bb686f33eea15835db13c0f9f997dcce275afb325102z" // testTxHash with a hex-invalid last character
		unknownTxHash    = "4262ac5a22fc593917a332fa80872ff88a57ccb211a3636a498b433149da4dee" // tx from another block
	)

	var (
		testBlock = identifier.Block{
			Index: &testHeight,
			Hash:  testBlockHash,
		}

		// corresponds to the block above
		testTx = identifier.Transaction{Hash: testTxHash}
	)

	tests := []struct {
		name string

		request request.Transaction

		checkErr assert.ErrorAssertionFunc
	}{
		{
			name:    "empty transaction request",
			request: request.Transaction{},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "missing blockchain name",
			request: request.Transaction{
				NetworkID: identifier.Network{
					Blockchain: "",
					Network:    dps.FlowLocalnet.String(),
				},
				BlockID:       testBlock,
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid blockchain name",
			request: request.Transaction{
				NetworkID: identifier.Network{
					Blockchain: invalidBlockchain,
					Network:    dps.FlowLocalnet.String(),
				},
				BlockID:       testBlock,
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidNetwork),
		},
		{
			name: "missing network name",
			request: request.Transaction{
				NetworkID: identifier.Network{
					Blockchain: dps.FlowBlockchain,
					Network:    "",
				},
				BlockID:       testBlock,
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid network name",
			request: request.Transaction{
				NetworkID: identifier.Network{
					Blockchain: dps.FlowBlockchain,
					Network:    invalidNetwork,
				},
				BlockID:       testBlock,
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidNetwork),
		},
		{
			name: "missing block height and hash",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: nil,
					Hash:  "",
				},
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid length of block id",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: &testHeight,
					Hash:  trimmedBlockHash,
				},
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "missing block height",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Hash: testBlockHash,
				},
				TransactionID: testTx,
			},
			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid block hash",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: &testHeight,
					Hash:  invalidBlockHash,
				},
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidBlock),
		},
		{
			name: "unknown block",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(lastHeight + 1),
					Hash:  mocks.GenericRosBlockID.Hash,
				},
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorUnknownBlock),
		},
		{
			name: "mismatched block height and hash",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID: identifier.Block{
					Index: getUint64P(44),
					Hash:  testBlockHash,
				},
				TransactionID: testTx,
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidBlock),
		},
		{
			name: "missing transaction id",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID:   testBlock,
				TransactionID: identifier.Transaction{
					Hash: "",
				},
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "missing transaction id",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID:   testBlock,
				TransactionID: identifier.Transaction{
					Hash: trimmedTxHash,
				},
			},

			checkErr: checkRosettaError(http.StatusBadRequest, configuration.ErrorInvalidFormat),
		},
		{
			name: "invalid transaction id",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID:   testBlock,
				TransactionID: identifier.Transaction{
					Hash: invalidTxHash,
				},
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorInvalidTransaction),
		},
		// TODO: Add test case for transaction with no events/transfers.
		//       See https://github.com/optakt/flow-dps/issues/452
		{
			name: "transaction missing from block",
			request: request.Transaction{
				NetworkID: defaultNetwork(),
				BlockID:   testBlock,
				TransactionID: identifier.Transaction{
					Hash: unknownTxHash,
				},
			},

			checkErr: checkRosettaError(http.StatusUnprocessableEntity, configuration.ErrorUnknownTransaction),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, ctx, err := setupRecorder(transactionEndpoint, test.request)
			require.NoError(t, err)

			err = data.Transaction(ctx)
			test.checkErr(t, err)
		})
	}
}

func TestAPI_TransactionHandlesMalformedRequest(t *testing.T) {

	db := setupDB(t)
	data := setupAPI(t, db)

	const (
		// network field is an integer instead of a string
		wrongFieldType = `
		{ 
			"network_identifier": { 
				"blockchain": "flow", 
				"network": 99
			}
		}`

		unclosedBracket = `
		{
			"network_identifier" : {
				"blockchain": "flow",
				"network": "flow-localnet"
			},
			"block_identifier": {
				"index": 105,
				"hash": "344368ba77ba47fb0a062dc8610c46cb2fe1539dbdedb0ba5fe8b46c629b0628"
			},
			"transaction_identifier": {
				"hash": "88419614bf6cda15586bb686f33eea15835db13c0f9f997dcce275afb325102a"
			}`

		validJSON = `
		{
			"network_identifier" : {
				"blockchain": "flow",
				"network": "flow-localnet"
			},
			"block_identifier": {
				"index": 105,
				"hash": "344368ba77ba47fb0a062dc8610c46cb2fe1539dbdedb0ba5fe8b46c629b0628"
			},
			"transaction_identifier": {
				"hash": "88419614bf6cda15586bb686f33eea15835db13c0f9f997dcce275afb325102a"
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

			_, ctx, err := setupRecorder(transactionEndpoint, test.payload, test.prepare)
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

func requestTransaction(header flow.Header, txID string) request.Transaction {
	return request.Transaction{
		NetworkID: defaultNetwork(),
		BlockID: identifier.Block{
			Index: &header.Height,
			Hash:  header.ID().String(),
		},
		TransactionID: identifier.Transaction{
			Hash: txID,
		},
	}
}
