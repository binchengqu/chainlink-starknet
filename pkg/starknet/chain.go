package starknet

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	"github.com/smartcontractkit/chainlink-starknet/pkg/starknet/db"

	// unused module to keep it go.mod and prevent ambiguous import
	_ "github.com/btcsuite/btcd/chaincfg/chainhash"
)

type ChainSet interface {
	types.Service

	Chain(ctx context.Context, id string) (Chain, error)
}

type Chain interface {
	types.Service

	Config() Config
	UpdateConfig(*db.ChainCfg)
}
