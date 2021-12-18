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

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/ziflex/lecho/v2"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk/client"

	api "github.com/optakt/flow-dps/api/dps"
	"github.com/optakt/flow-dps/codec/zbor"
	"github.com/optakt/flow-dps/models/dps"
	"github.com/optakt/flow-dps/service/invoker"

	rosetta "github.com/optakt/flow-dps-rosetta/api"
	"github.com/optakt/flow-dps-rosetta/service/configuration"
	"github.com/optakt/flow-dps-rosetta/service/converter"
	"github.com/optakt/flow-dps-rosetta/service/retriever"
	"github.com/optakt/flow-dps-rosetta/service/scripts"
	"github.com/optakt/flow-dps-rosetta/service/submitter"
	"github.com/optakt/flow-dps-rosetta/service/transactor"
	"github.com/optakt/flow-dps-rosetta/service/validator"
)

const (
	success = 0
	failure = 1
)

func main() {
	os.Exit(run())
}

func run() int {

	// Signal catching for clean shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Command line parameter initialization.
	var (
		flagDPS          string
		flagAccess       string
		flagCache        uint64
		flagLevel        string
		flagPort         uint16
		flagTransactions uint
		flagSmart        bool
	)

	pflag.StringVarP(&flagDPS, "dps-api", "a", "127.0.0.1:5005", "host address for GRPC API endpoint")
	pflag.StringVarP(&flagAccess, "access-api", "c", "access.canary.nodes.onflow.org:9000", "host address for Flow network's Access API endpoint")
	pflag.Uint64VarP(&flagCache, "cache", "e", 1_000_000_000, "maximum cache size for register reads in bytes")
	pflag.StringVarP(&flagLevel, "level", "l", "info", "log output level")
	pflag.Uint16VarP(&flagPort, "port", "p", 8080, "port to host Rosetta API on")
	pflag.UintVarP(&flagTransactions, "transaction-limit", "t", 200, "maximum amount of transactions to include in a block response")
	pflag.BoolVar(&flagSmart, "smart-status-codes", false, "enable smart non-500 HTTP status codes for Rosetta API errors")

	pflag.Parse()

	// Logger initialization.
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }
	log := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.DebugLevel)
	level, err := zerolog.ParseLevel(flagLevel)
	if err != nil {
		log.Error().Str("level", flagLevel).Err(err).Msg("could not parse log level")
		return failure
	}
	log = log.Level(level)
	elog := lecho.From(log)

	// Initialize codec.
	codec := zbor.NewCodec()

	// Initialize the DPS API client and wrap it for easy usage.
	conn, err := grpc.Dial(flagDPS, grpc.WithInsecure())
	if err != nil {
		log.Error().Str("api", flagDPS).Err(err).Msg("could not dial API host")
		return failure
	}
	defer conn.Close()
	dpsAPI := api.NewAPIClient(conn)
	index := api.IndexFromAPI(dpsAPI, codec)

	// Deduce chain ID from DPS API to configure parameters for script exec.
	first, err := index.First()
	if err != nil {
		log.Error().Err(err).Msg("could not get first height from DPS API")
		return failure
	}
	root, err := index.Header(first)
	if err != nil {
		log.Error().Uint64("first", first).Err(err).Msg("could not get root header from DPS API")
		return failure
	}
	params, ok := dps.FlowParams[root.ChainID]
	if !ok {
		log.Error().Str("chain", root.ChainID.String()).Msg("invalid chain ID for params")
		return failure
	}

	// Initialize the SDK client.
	if flagAccess == "" {
		log.Error().Msg("Flow Access API endpoint is missing")
		return failure
	}
	accessAPI, err := client.New(flagAccess, grpc.WithInsecure())
	if err != nil {
		log.Error().Str("address", flagAccess).Err(err).Msg("could not dial Flow Access API address")
		return failure
	}
	defer accessAPI.Close()

	// If smart status codes are enabled for the Rosetta API, we change the HTTP
	// status code constants here.
	if flagSmart {
		rosetta.EnableSmartCodes()
	}

	// Rosetta API initialization.
	config := configuration.New(params.ChainID)
	validate := validator.New(params, index, config)
	generate := scripts.NewGenerator(params)
	invoke, err := invoker.New(index, invoker.WithCacheSize(flagCache))
	if err != nil {
		log.Error().Err(err).Msg("could not initialize invoker")
		return failure
	}

	convert, err := converter.New(generate)
	if err != nil {
		log.Error().Err(err).Msg("could not generate transaction event types")
		return failure
	}

	retrieve := retriever.New(params, index, validate, generate, invoke, convert,
		retriever.WithTransactionLimit(flagTransactions),
	)
	dataCtrl := rosetta.NewData(config, retrieve, validate)

	submit := submitter.New(accessAPI)
	transact := transactor.New(validate, generate, invoke, submit)
	constructCtrl := rosetta.NewConstruction(config, transact, retrieve, validate)

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.Logger = elog
	server.Use(lecho.Middleware(lecho.Config{Logger: elog}))

	// This group contains all of the Rosetta Data API endpoints.
	server.POST("/network/list", dataCtrl.Networks)
	server.POST("/network/options", dataCtrl.Options)
	server.POST("/network/status", dataCtrl.Status)
	server.POST("/account/balance", dataCtrl.Balance)
	server.POST("/block", dataCtrl.Block)
	server.POST("/block/transaction", dataCtrl.Transaction)

	// This group contains all of the Rosetta Construction API endpoints.
	server.POST("/construction/preprocess", constructCtrl.Preprocess)
	server.POST("/construction/metadata", constructCtrl.Metadata)
	server.POST("/construction/payloads", constructCtrl.Payloads)
	server.POST("/construction/parse", constructCtrl.Parse)
	server.POST("/construction/combine", constructCtrl.Combine)
	server.POST("/construction/hash", constructCtrl.Hash)
	server.POST("/construction/submit", constructCtrl.Submit)

	// This section launches the main executing components in their own
	// goroutine, so they can run concurrently. Afterwards, we wait for an
	// interrupt signal in order to proceed with the next section.
	done := make(chan struct{})
	failed := make(chan struct{})
	go func() {
		log.Info().Msg("Flow Rosetta Server starting")
		err := server.Start(fmt.Sprint(":", flagPort))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn().Err(err).Msg("Flow Rosetta Server failed")
			close(failed)
		} else {
			close(done)
		}
		log.Info().Msg("Flow Rosetta Server stopped")
	}()

	select {
	case <-sig:
		log.Info().Msg("Flow Rosetta Server stopping")
	case <-done:
		log.Info().Msg("Flow Rosetta Server done")
	case <-failed:
		log.Warn().Msg("Flow Rosetta Server aborted")
		return failure
	}
	go func() {
		<-sig
		log.Warn().Msg("forcing exit")
		os.Exit(1)
	}()

	// The following code starts a shut down with a certain timeout and makes
	// sure that the main executing components are shutting down within the
	// allocated shutdown time. Otherwise, we will force the shutdown and log
	// an error. We then wait for shutdown on each component to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not shut down Rosetta API")
		return failure
	}

	return success
}
