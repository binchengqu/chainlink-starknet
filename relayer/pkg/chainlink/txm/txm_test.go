//go:build integration

package txm

import (
	"context"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	caigotypes "github.com/dontpanicdao/caigo/types"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/smartcontractkit/chainlink-starknet/ops"

	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/keys/mocks"
	txmmock "github.com/smartcontractkit/chainlink-starknet/relayer/pkg/chainlink/txm/mocks"
	"github.com/smartcontractkit/chainlink-starknet/relayer/pkg/starknet"
)

func TestIntegration_Txm(t *testing.T) {
	url := ops.SetupLocalStarknetNode(t)
	rawLocalKeys := ops.TestKeys(t, 2) // generate 2 keys

	// parse keys into expected format
	localKeys := map[string]keys.Key{}
	for _, k := range rawLocalKeys {
		key := keys.Raw(k).Key()
		account := "0x" + hex.EncodeToString(keys.PubKeyToAccount(key.PublicKey(), ops.DevnetClassHash, ops.DevnetSalt))
		localKeys[account] = key
	}

	// mock keystore
	ks := new(mocks.Keystore)
	ks.On("Get", mock.AnythingOfType("string")).Return(
		func(id string) keys.Key {
			return localKeys[id]
		},
		func(id string) error {

			_, ok := localKeys[id]
			if !ok {
				return errors.New("key does not exist")
			}
			return nil
		},
	)
	ks.On("GetAll").Return(maps.Values(localKeys), nil)

	lggr, err := logger.New()
	require.NoError(t, err)
	timeout := 10 * time.Second
	client, err := starknet.NewClient("devnet", url, lggr, &timeout)
	require.NoError(t, err)

	// test failing client in txm loop
	// first call - nonce syncer - success
	// second call - txm start loop - fail
	// third+more call - txm loop - success
	var count int
	var wg sync.WaitGroup
	wg.Add(3)

	// should be called twice
	getClient := func() (*starknet.Client, error) {
		count += 1
		wg.Done()
		if count == 2 {
			// test return not nil
			return &starknet.Client{}, errors.New("random test error")
		}

		return client, nil
	}

	// mock config to prevent import cycle
	cfg := txmmock.NewConfig(t)
	cfg.On("TxMaxBatchSize").Return(100)
	cfg.On("TxSendFrequency").Return(15 * time.Second)
	cfg.On("TxTimeout").Return(10 * time.Second)

	txm, err := New(lggr, ks, cfg, getClient)
	require.NoError(t, err)

	// ready fail if start not called
	require.Error(t, txm.Ready())

	// start txm + checks
	require.NoError(t, txm.Start(context.Background()))
	require.NoError(t, txm.Ready())

	for k := range localKeys {
		key := caigotypes.HexToHash(k)
		for i := 0; i < 5; i++ {
			require.NoError(t, txm.Enqueue(key, caigotypes.FunctionCall{
				ContractAddress:    key, // send to self
				EntryPointSelector: "get_nonce",
			}))
		}
	}
	time.Sleep(30 * time.Second)
	wg.Wait()

	// stop txm
	require.NoError(t, txm.Close())
	require.Error(t, txm.Ready())
}
