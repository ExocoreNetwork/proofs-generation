package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	commonutils "github.com/Layr-Labs/eigenpod-proofs-generation/common_utils"
)

// Setup function to create a real ProofServer with actual providers
func setupProofServer() *ProofServer {
	return NewProofServer(17000, "http://localhost:3500")
}

func TestGetValidatorProof(t *testing.T) {
	server := setupProofServer()

	// Test case
	req := &ValidatorProofRequest{
		Slot:           2300000,
		ValidatorIndex: 1647525,
	}

	// Act
	resp, err := server.GetValidatorProof(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// dump the response as json file for debugging
	// before write marshal response as json object, we should convert fields that are byte array to string
	proof := struct {
		StateRoot               string
		StateRootProof          []string
		ValidatorContainer      []string
		ValidatorContainerProof []string
	}{
		StateRoot:               "0x" + hex.EncodeToString(resp.StateRoot),
		StateRootProof:          commonutils.ConvertBytesToStrings(resp.StateRootProof),
		ValidatorContainer:      resp.ValidatorContainer,
		ValidatorContainerProof: commonutils.ConvertBytesToStrings(resp.ValidatorContainerProof),
	}
	proofjson, err := json.Marshal(proof)
	assert.NoError(t, err)
	os.WriteFile("validator_proof_test_1647525.json", proofjson, 0644)
}

func TestGetWithdrawalProof(t *testing.T) {
	server := setupProofServer()

	// Test case
	req := &WithdrawalProofRequest{
		StateSlot:      2300000,
		WithdrawalSlot: 2129382,
		ValidatorIndex: 1647525,
	}

	// Act
	resp, err := server.GetWithdrawalProof(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// dump the response as json file for debugging
	// before write marshal response as json object, we should convert fields that are byte array to string
	proof := struct {
		StateRoot                       string
		StateRootProof                  []string
		ValidatorContainer              []string
		ValidatorContainerProof         []string
		HistoricalSummaryBlockRoot      string
		HistoricalSummaryBlockRootProof []string
		SlotRoot                        string
		SlotRootProof                   []string
		TimestampRoot                   string
		TimestampRootProof              []string
		ExecutionPayloadRoot            string
		ExecutionPayloadRootProof       []string
		WithdrawalContainer             []string
		WithdrawalContainerProof        []string
		HistoricalSummaryIndex          uint64
		BlockRootIndex                  uint64
		WithdrawalIndexWithinBlock      uint64
	}{
		StateRoot:                       "0x" + hex.EncodeToString(resp.StateRoot),
		StateRootProof:                  commonutils.ConvertBytesToStrings(resp.StateRootProof),
		ValidatorContainer:              resp.ValidatorContainer,
		ValidatorContainerProof:         commonutils.ConvertBytesToStrings(resp.ValidatorContainerProof),
		HistoricalSummaryBlockRoot:      "0x" + hex.EncodeToString(resp.HistoricalSummaryBlockRoot),
		HistoricalSummaryBlockRootProof: commonutils.ConvertBytesToStrings(resp.HistoricalSummaryBlockRootProof),
		SlotRoot:                        "0x" + hex.EncodeToString(resp.SlotRoot),
		SlotRootProof:                   commonutils.ConvertBytesToStrings(resp.SlotRootProof),
		TimestampRoot:                   "0x" + hex.EncodeToString(resp.TimestampRoot),
		TimestampRootProof:              commonutils.ConvertBytesToStrings(resp.TimestampRootProof),
		ExecutionPayloadRoot:            "0x" + hex.EncodeToString(resp.ExecutionPayloadRoot),
		ExecutionPayloadRootProof:       commonutils.ConvertBytesToStrings(resp.ExecutionPayloadRootProof),
		WithdrawalContainer:             commonutils.ConvertBytesToStrings(resp.WithdrawalContainer),
		WithdrawalContainerProof:        commonutils.ConvertBytesToStrings(resp.WithdrawalContainerProof),
		HistoricalSummaryIndex:          resp.HistoricalSummaryIndex,
		BlockRootIndex:                  resp.BlockRootIndex,
		WithdrawalIndexWithinBlock:      resp.WithdrawalIndexWithinBlock,
	}
	proofjson, err := json.Marshal(proof)
	assert.NoError(t, err)
	os.WriteFile("withdrawal_proof_test_1647525.json", proofjson, 0644)
}
