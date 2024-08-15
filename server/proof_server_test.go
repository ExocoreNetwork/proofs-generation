package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Setup function to create a real ProofServer with actual providers
func setupProofServer() *ProofServer {
	return NewProofServer(17000, "http://localhost:8545")
}

func TestGetValidatorProof(t *testing.T) {
	server := setupProofServer()

	// Test case
	req := &ValidatorProofRequest{
		Slot:           123,
		ValidatorIndex: 456,
	}

	// Act
	resp, err := server.GetValidatorProof(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// Add more specific assertions based on expected response
	// For example:
	// assert.NotEmpty(t, resp.Proof)
	// assert.Equal(t, req.Slot, resp.Slot)
	// assert.Equal(t, req.ValidatorIndex, resp.ValidatorIndex)
}

func TestGetWithdrawalProof(t *testing.T) {
	server := setupProofServer()

	// Test case
	req := &WithdrawalProofRequest{
		StateSlot:      123,
		WithdrawalSlot: 456,
		ValidatorIndex: 789,
	}

	// Act
	resp, err := server.GetWithdrawalProof(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// Add more specific assertions based on expected response
	// For example:
	// assert.NotEmpty(t, resp.Proof)
	// assert.Equal(t, req.StateSlot, resp.StateSlot)
	// assert.Equal(t, req.WithdrawalSlot, resp.WithdrawalSlot)
	// assert.Equal(t, req.ValidatorIndex, resp.ValidatorIndex)
}

// Add more test functions as needed
