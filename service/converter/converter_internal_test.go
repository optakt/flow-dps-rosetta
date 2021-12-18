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

package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/encoding/json"
	"github.com/onflow/cadence/runtime/tests/utils"
	"github.com/onflow/flow-go/model/flow"

	"github.com/optakt/flow-dps-rosetta/service/identifier"
	"github.com/optakt/flow-dps-rosetta/service/object"
	"github.com/optakt/flow-dps-rosetta/service/retriever"
	"github.com/optakt/flow-dps-rosetta/testing/mocks"
	"github.com/optakt/flow-dps/models/dps"
)

func TestNew(t *testing.T) {
	t.Run("nominal case", func(t *testing.T) {
		generator := mocks.BaselineGenerator(t)
		generator.TokensDepositedFunc = func(symbol string) (string, error) {
			assert.Equal(t, dps.FlowSymbol, symbol)
			return string(mocks.GenericEventType(0)), nil
		}
		generator.TokensWithdrawnFunc = func(symbol string) (string, error) {
			assert.Equal(t, dps.FlowSymbol, symbol)
			return string(mocks.GenericEventType(1)), nil
		}

		cvt, err := New(generator)

		require.NoError(t, err)
		assert.Equal(t, cvt.deposit, mocks.GenericEventType(0))
		assert.Equal(t, cvt.withdrawal, mocks.GenericEventType(1))
	})

	t.Run("handles generator failure for deposit event type", func(t *testing.T) {
		generator := mocks.BaselineGenerator(t)
		generator.TokensDepositedFunc = func(symbol string) (string, error) {
			return "", mocks.GenericError
		}

		cvt, err := New(generator)

		assert.Error(t, err)
		assert.Nil(t, cvt)
	})

	t.Run("handles generator failure for withdrawal event type", func(t *testing.T) {
		generator := mocks.BaselineGenerator(t)
		generator.TokensWithdrawnFunc = func(symbol string) (string, error) {
			return "", mocks.GenericError
		}

		cvt, err := New(generator)

		assert.Error(t, err)
		assert.Nil(t, cvt)
	})
}

func TestConverter_EventToOperation(t *testing.T) {
	depositType := &cadence.EventType{
		Location:            utils.TestLocation,
		QualifiedIdentifier: string(mocks.GenericEventType(0)),
		Fields: []cadence.Field{
			{
				Identifier: "amount",
				Type:       cadence.UInt64Type{},
			},
			{
				Identifier: "address",
				Type:       cadence.AddressType{},
			},
		},
	}
	depositEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewUInt64(42),
			cadence.NewAddress([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		},
	).WithType(depositType)
	depositEventPayload := json.MustEncode(depositEvent)

	withdrawalType := &cadence.EventType{
		Location:            utils.TestLocation,
		QualifiedIdentifier: string(mocks.GenericEventType(1)),
		Fields: []cadence.Field{
			{
				Identifier: "amount",
				Type:       cadence.UInt64Type{},
			},
			{
				Identifier: "address",
				Type:       cadence.AddressType{},
			},
		},
	}
	withdrawalEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewUInt64(42),
			cadence.NewAddress([8]byte{2, 3, 4, 5, 6, 7, 8, 9}),
		},
	).WithType(withdrawalType)
	withdrawalEventPayload := json.MustEncode(withdrawalEvent)

	depositNetIndex := uint(1)
	testDepositOp := object.Operation{
		ID: identifier.Operation{
			NetworkIndex: &depositNetIndex,
		},
		Type:   dps.OperationTransfer,
		Status: dps.StatusCompleted,
		AccountID: identifier.Account{
			Address: "0102030405060708",
		},
		Amount: object.Amount{
			Value: "42",
			Currency: identifier.Currency{
				Symbol:   dps.FlowSymbol,
				Decimals: dps.FlowDecimals,
			},
		},
	}
	withdrawalNetIndex := uint(2)
	testWithdrawalOp := object.Operation{
		ID: identifier.Operation{
			NetworkIndex: &withdrawalNetIndex,
		},
		Type:   dps.OperationTransfer,
		Status: dps.StatusCompleted,
		AccountID: identifier.Account{
			Address: "0203040506070809",
		},
		Amount: object.Amount{
			Value: "-42",
			Currency: identifier.Currency{
				Symbol:   dps.FlowSymbol,
				Decimals: dps.FlowDecimals,
			},
		},
	}

	id, err := flow.HexStringToIdentifier("a4c4194eae1a2dd0de4f4d51a884db4255bf265a40ddd98477a1d60ef45909ec")
	require.NoError(t, err)

	threeFieldsType := &cadence.EventType{
		Location:            utils.TestLocation,
		QualifiedIdentifier: "test",
		Fields: []cadence.Field{
			{
				Identifier: "testField1",
				Type:       cadence.UInt64Type{},
			},
			{
				Identifier: "testField2",
				Type:       cadence.UInt64Type{},
			},
			{
				Identifier: "testField3",
				Type:       cadence.UInt64Type{},
			},
		},
	}
	threeFieldsEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewUInt64(42),
			cadence.NewUInt64(42),
			cadence.NewUInt64(42),
		},
	).WithType(threeFieldsType)
	threeFieldsEventPayload := json.MustEncode(threeFieldsEvent)

	missingAmountEventType := &cadence.EventType{
		Location:            utils.TestLocation,
		QualifiedIdentifier: "test",
		Fields: []cadence.Field{
			{
				Identifier: "address",
				Type:       cadence.AddressType{},
			},
			{
				Identifier: "testField",
				Type:       cadence.AddressType{},
			},
		},
	}
	missingAmountEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewAddress([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			cadence.NewAddress([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
		},
	).WithType(missingAmountEventType)
	missingAmountEventPayload := json.MustEncode(missingAmountEvent)

	missingAddressEventType := &cadence.EventType{
		Location:            utils.TestLocation,
		QualifiedIdentifier: "test",
		Fields: []cadence.Field{
			{
				Identifier: "amount",
				Type:       cadence.UInt64Type{},
			},
			{
				Identifier: "amount",
				Type:       cadence.UInt64Type{},
			},
		},
	}
	missingAddressEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewUInt64(42),
			cadence.NewUInt64(42),
		},
	).WithType(missingAddressEventType)
	missingAddressEventPayload := json.MustEncode(missingAddressEvent)

	nilAddressEvent := cadence.NewEvent(
		[]cadence.Value{
			cadence.NewUInt64(42),
			cadence.NewOptional(nil),
		},
	).WithType(withdrawalType)
	nilAddressPayload := json.MustEncode(nilAddressEvent)

	tests := []struct {
		name string

		event flow.Event

		wantErr       assert.ErrorAssertionFunc
		wantSentinel  error
		wantOperation *object.Operation
	}{
		{
			name: "nominal case with deposit event",

			event: flow.Event{
				TransactionID: id,
				Type:          mocks.GenericEventType(0),
				Payload:       depositEventPayload,
				EventIndex:    1,
			},

			wantErr:       assert.NoError,
			wantOperation: &testDepositOp,
		},
		{
			name: "nominal case with withdrawal event",

			event: flow.Event{
				TransactionID: id,
				Type:          mocks.GenericEventType(1),
				Payload:       withdrawalEventPayload,
				EventIndex:    2,
			},

			wantErr:       assert.NoError,
			wantOperation: &testWithdrawalOp,
		},
		{
			name: "unsupported event type",

			event: flow.Event{
				TransactionID: id,
				Type:          flow.EventType("irrelevant"),
				Payload:       withdrawalEventPayload,
			},

			wantErr:      assert.Error,
			wantSentinel: retriever.ErrNotSupported,
		},
		{
			name: "wrong amount of fields",

			event: flow.Event{
				Type:    mocks.GenericEventType(0),
				Payload: threeFieldsEventPayload,
			},

			wantErr: assert.Error,
		},
		{
			name: "missing amount field",

			event: flow.Event{
				Type:    mocks.GenericEventType(0),
				Payload: missingAmountEventPayload,
			},

			wantErr: assert.Error,
		},
		{
			name: "missing address field",

			event: flow.Event{
				Type:    mocks.GenericEventType(0),
				Payload: missingAddressEventPayload,
			},

			wantErr: assert.Error,
		},
		{
			name: "nil address field",

			event: flow.Event{
				TransactionID: id,
				Type:          mocks.GenericEventType(0),
				Payload:       nilAddressPayload,
			},

			wantErr:      assert.Error,
			wantSentinel: retriever.ErrNoAddress,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cvt := &Converter{
				deposit:    mocks.GenericEventType(0),
				withdrawal: mocks.GenericEventType(1),
			}

			got, err := cvt.EventToOperation(test.event)

			test.wantErr(t, err)
			if test.wantSentinel != nil {
				assert.ErrorIs(t, err, test.wantSentinel)
			}

			assert.Equal(t, test.wantOperation, got)
		})
	}
}
