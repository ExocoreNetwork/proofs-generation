package main

import (
	"context"
	"strconv"

	eigenpodproofs "github.com/Layr-Labs/eigenpod-proofs-generation"
	"github.com/Layr-Labs/eigenpod-proofs-generation/beacon"
	commonutils "github.com/Layr-Labs/eigenpod-proofs-generation/common_utils"
	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ProofServer struct {
	chainId             uint64
	blockHeaderProvider eth2client.BeaconBlockHeadersProvider
	stateProvider       eth2client.BeaconStateProvider
	blockProvider       eth2client.SignedBeaconBlockProvider
}

func NewProofServer(chainId uint64, provider string) (server *ProofServer) {
	// Provide a cancellable context to the creation function.
	ctx, _ := context.WithCancel(context.Background())
	client, err := http.New(ctx,
		// WithAddress supplies the address of the beacon node, as a URL.
		http.WithAddress(provider),
		// LogLevel supplies the level of logging to carry out.
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		panic(err)
	}

	if provider, isProvider := client.(eth2client.BeaconBlockHeadersProvider); isProvider {
		server.blockHeaderProvider = provider
	} else {
		panic("not a beacon block header provider")
	}

	if provider, isProvider := client.(eth2client.BeaconStateProvider); isProvider {
		server.stateProvider = provider
	} else {
		panic("not a beacon state provider")
	}

	if provider, isProvider := client.(eth2client.SignedBeaconBlockProvider); isProvider {
		server.blockProvider = provider
	} else {
		panic("not a beacon block provider")
	}

	server.chainId = chainId

	return server
}

func (s *ProofServer) GetValidatorProof(ctx context.Context, req *ValidatorProofRequest) (*ValidatorProofResponse, error) {
	// TODO: check slot is after deneb fork

	var beaconBlockHeader *phase0.BeaconBlockHeader
	var versionedState *spec.VersionedBeaconState

	blockHeaderResponse, err := s.blockHeaderProvider.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{Block: strconv.FormatUint(req.Slot, 10)})
	if err != nil {
		return nil, err
	}
	beaconBlockHeader = blockHeaderResponse.Data.Header.Message

	beaconStateResponse, err := s.stateProvider.BeaconState(ctx, &api.BeaconStateOpts{State: strconv.FormatUint(req.Slot, 10)})
	if err != nil {
		return nil, err
	}
	versionedState = beaconStateResponse.Data

	epp, err := eigenpodproofs.NewEigenPodProofs(s.chainId, 1000)
	if err != nil {
		log.Debug().AnErr("Error creating EPP object", err)
		return nil, err
	}

	stateRootProof, validatorContainerProof, err := eigenpodproofs.ProveValidatorFields(epp, beaconBlockHeader, versionedState, req.ValidatorIndex)
	if err != nil {
		log.Debug().AnErr("Error with ProveValidatorFields", err)
		return nil, err
	}

	return &ValidatorProofResponse{
		StateRoot:               beaconBlockHeader.StateRoot[:],
		StateRootProof:          stateRootProof.SlotRootProof.ToBytesSlice(),
		ValidatorContainer:      commonutils.GetValidatorFields(versionedState.Deneb.Validators[req.ValidatorIndex]),
		ValidatorContainerProof: validatorContainerProof.ToBytesSlice(),
	}, nil
}

func (s *ProofServer) GetWithdrawalProof(ctx context.Context, req *WithdrawalProofRequest) (*WithdrawalProofResponse, error) {
	var oracleBlockHeader *phase0.BeaconBlockHeader
	var oracleState *spec.VersionedBeaconState
	var withdrawalBlock *spec.VersionedSignedBeaconBlock
	var completeTargetBlockRootsGroup []phase0.Root
	var completeTargetBlockRootsGroupSlot uint64
	var targetBlockRootsGroupSummaryIndex uint64
	var withdrawalBlockRootIndexInGroup uint64
	var historicalSummaryBlockRoot []byte
	var withdrawalIndex uint64

	epp, err := eigenpodproofs.NewEigenPodProofs(s.chainId, 1000)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error creating EPP object", err)
		return nil, err
	}

	targetBlockRootsGroupSummaryIndex, withdrawalBlockRootIndexInGroup, completeTargetBlockRootsGroupSlot, err = epp.GetWithdrawalProofParams(req.StateSlot, req.WithdrawalSlot)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error getting withdrawal proof params", err)
		return nil, err
	}

	blockHeaderResponse, err := s.blockHeaderProvider.BeaconBlockHeader(ctx, &api.BeaconBlockHeaderOpts{Block: strconv.FormatUint(req.StateSlot, 10)})
	if err != nil {
		return nil, err
	}
	oracleBlockHeader = blockHeaderResponse.Data.Header.Message

	beaconStateResponse, err := s.stateProvider.BeaconState(ctx, &api.BeaconStateOpts{State: strconv.FormatUint(req.StateSlot, 10)})
	if err != nil {
		return nil, err
	}
	oracleState = beaconStateResponse.Data

	blockResponse, err := s.blockProvider.SignedBeaconBlock(ctx, &api.SignedBeaconBlockOpts{Block: strconv.FormatUint(req.WithdrawalSlot, 10)})
	if err != nil {
		return nil, err
	}
	withdrawalBlock = blockResponse.Data

	beaconStateResponse, err = s.stateProvider.BeaconState(ctx, &api.BeaconStateOpts{State: strconv.FormatUint(completeTargetBlockRootsGroupSlot, 10)})
	if err != nil {
		return nil, err
	}
	completeTargetBlockRootsGroup = beaconStateResponse.Data.Deneb.BlockRoots

	oracleStateContainerTopLevelRoots, err := epp.ComputeBeaconStateTopLevelRoots(oracleState)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with ComputeBeaconStateTopLevelRoots", err)
		return nil, err
	}

	withdrawalContainerProof, withdrawalContainer, err := epp.ProveWithdrawal(oracleBlockHeader, oracleState, oracleStateContainerTopLevelRoots, completeTargetBlockRootsGroup, withdrawalBlock, req.ValidatorIndex)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with ProveWithdrawal", err)
		return nil, err
	}

	var withdrawalContainerBytes [][]byte
	for _, withdrawal := range withdrawalContainer {
		withdrawalContainerBytes = append(withdrawalContainerBytes, withdrawal[:])
	}

	stateRootProofAgainstBlockHeader, err := beacon.ProveStateRootAgainstBlockHeader(oracleBlockHeader)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with ProveStateRootAgainstBlockHeader", err)
		return nil, err
	}

	slotProofAgainstBlockHeader, err := beacon.ProveSlotAgainstBlockHeader(oracleBlockHeader)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with ProveSlotAgainstBlockHeader", err)
		return nil, err
	}

	validators, err := oracleState.Validators()
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with get validators", err)
		return nil, err
	}
	validatorContainerProof, err := epp.ProveValidatorAgainstBeaconState(oracleStateContainerTopLevelRoots, phase0.Slot(req.StateSlot), validators, req.ValidatorIndex)
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with ProveValidatorAgainstBeaconState", err)
		return nil, err
	}

	root, err := withdrawalBlock.Root()
	if err != nil {
		log.Debug().AnErr("GenerateWithdrawalFieldsProof: error with get withdrawal block root", err)
		return nil, err
	}
	historicalSummaryBlockRoot = root[:]

	withdrawalIndex = withdrawalContainerProof.WithdrawalIndex

	return &WithdrawalProofResponse{
		StateRoot:                       oracleBlockHeader.StateRoot[:],
		StateRootProof:                  stateRootProofAgainstBlockHeader.ToBytesSlice(),
		ValidatorContainer:              commonutils.GetValidatorFields(validators[req.ValidatorIndex]),
		ValidatorContainerProof:         validatorContainerProof.ToBytesSlice(),
		HistoricalSummaryBlockRoot:      historicalSummaryBlockRoot,
		HistoricalSummaryBlockRootProof: withdrawalContainerProof.HistoricalSummaryBlockRootProof.ToBytesSlice(),
		SlotRoot:                        withdrawalContainerProof.SlotRoot[:],
		SlotRootProof:                   slotProofAgainstBlockHeader.ToBytesSlice(),
		TimestampRoot:                   withdrawalContainerProof.TimestampRoot[:],
		TimestampRootProof:              withdrawalContainerProof.TimestampProof.ToBytesSlice(),
		ExecutionPayloadRoot:            withdrawalContainerProof.ExecutionPayloadRoot[:],
		ExecutionPayloadRootProof:       withdrawalContainerProof.ExecutionPayloadProof.ToBytesSlice(),
		WithdrawalContainer:             withdrawalContainerBytes,
		WithdrawalContainerProof:        withdrawalContainerProof.WithdrawalProof.ToBytesSlice(),
		HistoricalSummaryIndex:          targetBlockRootsGroupSummaryIndex,
		BlockRootIndex:                  withdrawalBlockRootIndexInGroup,
		WithdrawalIndexWithinBlock:      withdrawalIndex,
	}, err
}
