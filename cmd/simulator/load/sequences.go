// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package load

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/subnet-evm/cmd/simulator/config"
	"github.com/ava-labs/subnet-evm/cmd/simulator/txs"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/ethclient"
	"github.com/ava-labs/subnet-evm/params"
	"github.com/ava-labs/subnet-evm/warp"
)

type TxSequenceGetter func(
	ctx context.Context, config config.Config, chainID *big.Int,
	pks []*ecdsa.PrivateKey, client ethclient.Client,
) ([]txs.TxSequence[*types.Transaction], error)

// func GetEVMTxSequences(
// 	ctx context.Context, config config.Config, chainID *big.Int,
// 	pks []*ecdsa.PrivateKey, client ethclient.Client,
//	tracker *awmTimeTracker,
// ) ([]txs.TxSequence[TrackableTx], error) {
// 	bigGwei := big.NewInt(params.GWei)
// 	gasTipCap := new(big.Int).Mul(bigGwei, big.NewInt(config.MaxTipCap))
// 	gasFeeCap := new(big.Int).Mul(bigGwei, big.NewInt(config.MaxFeeCap))
//
// 	// Normal EVM txs
// 	signer := types.LatestSignerForChainID(chainID)
// 	txGenerator := func(key *ecdsa.PrivateKey, nonce uint64) (*types.Transaction, error) {
// 		addr := ethcrypto.PubkeyToAddress(key.PublicKey)
// 		tx, err := types.SignNewTx(key, signer, &types.DynamicFeeTx{
// 			ChainID:   chainID,
// 			Nonce:     nonce,
// 			GasTipCap: gasTipCap,
// 			GasFeeCap: gasFeeCap,
// 			Gas:       params.TxGas,
// 			To:        &addr,
// 			Data:      nil,
// 			Value:     common.Big0,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 		return tx, nil
// 	}
// 	return txs.GenerateTxSequences(ctx, txGenerator, client, pks, config.TxsPerWorker)
// }

func GetWarpSendTxSequences(
	ctx context.Context, config config.Config, chainID *big.Int,
	pks []*ecdsa.PrivateKey, startingNonces []uint64,
) ([]txs.TxSequence[*AwmTx], error) {
	// TODO: pass through beginning nonces instead of client here.
	bigGwei := big.NewInt(params.GWei)
	gasTipCap := new(big.Int).Mul(bigGwei, big.NewInt(config.MaxTipCap))
	gasFeeCap := new(big.Int).Mul(bigGwei, big.NewInt(config.MaxFeeCap))

	subnetBStr := os.Getenv("SUBNET_B") // TODO: pass properly through config
	subnetB, err := ids.FromString(subnetBStr)
	if err != nil {
		return nil, err
	}
	txGenerator := MkSendWarpTxGenerator(chainID, subnetB, gasFeeCap, gasTipCap)
	return txs.GenerateTxSequences(ctx, txGenerator, pks, startingNonces, config.TxsPerWorker)
}

func GetWarpReceiveTxSequences(
	ctx context.Context, config config.Config, chainID *big.Int,
	pks []*ecdsa.PrivateKey, startingNonces []uint64,
) ([]txs.TxSequence[*AwmTx], error) {
	ch := make(chan warpSignature) // channel for incoming signatures
	// We will need to aggregate signatures for messages that are sent on
	// subnet A. So we will subscribe to the subnet A's accepted logs.
	// TODO: fix how we get ethclients for subnet A here.
	endpointsStr := os.Getenv("RPC_ENDPOINTS_SUBNET_A")
	endpoints := strings.Split(endpointsStr, ",")
	clients := make([]ethclient.Client, len(endpoints))
	for i, clientURI := range endpoints {
		client, err := ethclient.Dial(clientURI)
		if err != nil {
			return nil, fmt.Errorf("failed to dial client at %s: %w", clientURI, err)
		}
		clients[i] = client
	}
	for i, client := range clients {
		// TODO: remove this hack
		endpoint := endpoints[i] // %s/ext/bc/%s/rpc
		split := strings.Split(endpoint, "/")
		chain := split[len(split)-2]
		uri := strings.Join(split[:len(split)-4], "/")

		warpClient, err := warp.NewWarpClient(uri, chain)
		if err != nil {
			return nil, err
		}
		// TODO: this index should correspond to P-Chain validator index
		// TODO: properly shutdown warp clients
		_ = NewWarpRelayClient(ctx, client, warpClient, ch, i)
	}

	threshold := uint64(4) // TODO: should not be hardcoded
	// TODO: should not be hardcoded like this
	expectedMessages := int(config.TxsPerWorker) * config.Workers
	warpRelay := NewWarpRelay(ctx, threshold, ch, expectedMessages)
	// Each worker will listen for signed warp messages that are
	// ready to be issued
	txSequences := make([]txs.TxSequence[*AwmTx], config.Workers)
	for i := 0; i < config.Workers; i++ {
		txSequences[i] = NewWarpRelayTxSequence(ctx, warpRelay.signedMessages, chainID, pks[i], startingNonces[i])
	}
	return txSequences, nil
}