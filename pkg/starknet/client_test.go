package starknet

import (
	"context"
	"testing"

	"github.com/dontpanicdao/caigo/gateway"
	"github.com/stretchr/testify/assert"

	// "github.com/smartcontractkit/chainlink-starknet/pkg/starknet/ocr2"
	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

func TestGatewayClient(t *testing.T) {
	// todo: adjust for e2e tests
	chainID := gateway.GOERLI_ID
	ocr2ContractAddress := "0x756ce9ca3dff7ee1037e712fb9662be13b5dcfc0660b97d266298733e1196b"
	lggr := logger.Test(t)

	client, err := NewClient(chainID, lggr)
	assert.NoError(t, err)

	t.Run("get chain id", func(t *testing.T) {
		id, err := client.ChainID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, id, chainID)
	})

	t.Run("get block height", func(t *testing.T) {
		_, err := client.LatestBlockHeight(context.Background())
		assert.NoError(t, err)
	})

	t.Run("get billing details", func(t *testing.T) {
		_, err := client.OCR2BillingDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)
	})

	t.Run("get latest config details", func(t *testing.T) {
		details, err := client.OCR2LatestConfigDetails(context.Background(), ocr2ContractAddress)
		assert.NoError(t, err)

		_, err = client.OCR2LatestConfig(context.Background(), ocr2ContractAddress, details.Block)
		assert.NoError(t, err)
	})
}
