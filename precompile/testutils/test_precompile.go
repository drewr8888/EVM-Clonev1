// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package testutils

import (
	"math/big"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/subnet-evm/commontype"
	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/precompile/modules"
	"github.com/ava-labs/subnet-evm/precompile/precompileconfig"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var DefaultChainConfig = contract.NewMockChainConfig(commontype.ValidTestFeeConfig, false, utils.NewUint64(0))

// PrecompileTest is a test case for a precompile
type PrecompileTest struct {
	// Caller is the address of the precompile caller
	Caller common.Address
	// Input the raw input bytes to the precompile
	Input []byte
	// InputFn is a function that returns the raw input bytes to the precompile
	// If specified, Input will be ignored.
	InputFn func(t testing.TB) []byte
	// SuppliedGas is the amount of gas supplied to the precompile
	SuppliedGas uint64
	// ReadOnly is whether the precompile should be called in read only
	// mode. If true, the precompile should not modify the state.
	ReadOnly bool
	// Config is the config to use for the precompile
	// It should be the same precompile config that is used in the
	// precompile's configurator.
	// If nil, Configure will not be called.
	Config precompileconfig.Config
	// BeforeHook is called before the precompile is called.
	BeforeHook func(t testing.TB, state contract.StateDB)
	// AfterHook is called after the precompile is called.
	AfterHook func(t testing.TB, state contract.StateDB)
	// ExpectedRes is the expected raw byte result returned by the precompile
	ExpectedRes []byte
	// ExpectedErr is the expected error returned by the precompile
	ExpectedErr string
	// BlockNumber is the block number to use for the precompile's block context
	BlockNumber int64
	// ChainConfig is the chain config to use for the precompile's block context
	// If nil, the default chain config will be used.
	ChainConfig contract.ChainConfig
}

type PrecompileRunparams struct {
	AccessibleState contract.AccessibleState
	Caller          common.Address
	ContractAddress common.Address
	Input           []byte
	SuppliedGas     uint64
	ReadOnly        bool
}

func (test PrecompileTest) Run(t *testing.T, module modules.Module, state contract.StateDB) {
	runParams := test.setup(t, module, state)

	if runParams.Input != nil {
		ret, remainingGas, err := module.Contract.Run(runParams.AccessibleState, runParams.Caller, runParams.ContractAddress, runParams.Input, runParams.SuppliedGas, runParams.ReadOnly)
		if len(test.ExpectedErr) != 0 {
			require.ErrorContains(t, err, test.ExpectedErr)
		} else {
			require.NoError(t, err)
		}
		require.Equal(t, uint64(0), remainingGas)
		require.Equal(t, test.ExpectedRes, ret)
	}

	if test.AfterHook != nil {
		test.AfterHook(t, state)
	}
}

func (test PrecompileTest) setup(t testing.TB, module modules.Module, state contract.StateDB) PrecompileRunparams {
	t.Helper()
	contractAddress := module.Address

	if test.BeforeHook != nil {
		test.BeforeHook(t, state)
	}

	blockContext := contract.NewMockBlockContext(big.NewInt(test.BlockNumber), 0)
	chainConfig := test.ChainConfig
	if chainConfig == nil {
		// DUpgrade is activated by default
		chainConfig = DefaultChainConfig
	}

	accesibleState := contract.NewMockAccessibleState(state, blockContext, snow.DefaultContextTest(), chainConfig)

	if test.Config != nil {
		err := module.Configure(chainConfig, test.Config, state, blockContext)
		require.NoError(t, err)
	}

	input := test.Input
	if test.InputFn != nil {
		input = test.InputFn(t)
	}

	return PrecompileRunparams{
		AccessibleState: accesibleState,
		Caller:          test.Caller,
		ContractAddress: contractAddress,
		Input:           input,
		SuppliedGas:     test.SuppliedGas,
		ReadOnly:        test.ReadOnly,
	}
}

func (test PrecompileTest) Bench(b *testing.B, module modules.Module, state contract.StateDB) {
	runParams := test.setup(b, module, state)

	if runParams.Input == nil {
		b.Skip("Skipping precompile benchmark due to nil input (used for configuration tests)")
	}

	stateDB := runParams.AccessibleState.GetStateDB()
	snapshot := stateDB.Snapshot()

	ret, remainingGas, err := module.Contract.Run(runParams.AccessibleState, runParams.Caller, runParams.ContractAddress, runParams.Input, runParams.SuppliedGas, runParams.ReadOnly)
	if len(test.ExpectedErr) != 0 {
		require.ErrorContains(b, err, test.ExpectedErr)
	} else {
		require.NoError(b, err)
	}
	require.Equal(b, uint64(0), remainingGas)
	require.Equal(b, test.ExpectedRes, ret)

	if test.AfterHook != nil {
		test.AfterHook(b, state)
	}

	b.ReportAllocs()
	start := time.Now()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Revert to the previous snapshot and take a new snapshot, so we can reset the state after execution
		stateDB.RevertToSnapshot(snapshot)
		snapshot = stateDB.Snapshot()

		// Ignore return values for benchmark
		_, _, _ = module.Contract.Run(runParams.AccessibleState, runParams.Caller, runParams.ContractAddress, runParams.Input, runParams.SuppliedGas, runParams.ReadOnly)
	}
	b.StopTimer()

	elapsed := uint64(time.Since(start))
	if elapsed < 1 {
		elapsed = 1
	}
	gasUsed := runParams.SuppliedGas * uint64(b.N)
	b.ReportMetric(float64(runParams.SuppliedGas), "gas/op")
	// Keep it as uint64, multiply 100 to get two digit float later
	mgasps := (100 * 1000 * gasUsed) / elapsed
	b.ReportMetric(float64(mgasps)/100, "mgas/s")

	// Execute the test one final time to ensure that if our RevertToSnapshot logic breaks such that each run is actually failing or resulting in unexpected behavior
	// the benchmark should catch the error here.
	stateDB.RevertToSnapshot(snapshot)
	ret, remainingGas, err = module.Contract.Run(runParams.AccessibleState, runParams.Caller, runParams.ContractAddress, runParams.Input, runParams.SuppliedGas, runParams.ReadOnly)
	if len(test.ExpectedErr) != 0 {
		require.ErrorContains(b, err, test.ExpectedErr)
	} else {
		require.NoError(b, err)
	}
	require.Equal(b, uint64(0), remainingGas)
	require.Equal(b, test.ExpectedRes, ret)

	if test.AfterHook != nil {
		test.AfterHook(b, state)
	}
}

func RunPrecompileTests(t *testing.T, module modules.Module, newStateDB func(t testing.TB) contract.StateDB, contractTests map[string]PrecompileTest) {
	t.Helper()

	for name, test := range contractTests {
		t.Run(name, func(t *testing.T) {
			test.Run(t, module, newStateDB(t))
		})
	}
}
