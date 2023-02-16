// Code generated
// This file is a generated precompile contract config with stubbed abstract functions.
// The file is generated by a template. Please inspect every code and comment in this file before use.

// There are some must-be-done changes waiting in the file. Each area requiring you to add your code is marked with CUSTOM CODE to make them easy to find and modify.
// Additionally there are other files you need to edit to activate your precompile.
// These areas are highlighted with comments "ADD YOUR PRECOMPILE HERE".
// For testing take a look at other precompile tests in core/stateful_precompile_test.go and config_test.go in other precompile folders.

/* General guidelines for precompile development:
1- Read the comment and set a suitable contract address in generated contract.go. E.g:
	ContractAddress = common.HexToAddress("ASUITABLEHEXADDRESS")
2- It is recommended to only modify code in the highlighted areas marked with "CUSTOM CODE STARTS HERE". Modifying code outside of these areas should be done with caution and with a deep understanding of how these changes may impact the EVM.
Typically, custom codes are required in only those areas.
3- Set gas costs in generated contract.go
4- Register your precompile module in params/precompile_modules.go
5- Add your config unit tests under generated package config_test.go
6- Add your contract unit tests under tests/statefulprecompiles/{precompilename}_test.go
7- Add your solidity interface and test contract to contract-examples/contracts
8- Write solidity tests for your precompile in contract-examples/test
9- Create your genesis with your precompile enabled in tests/e2e/genesis/
10- Create e2e test for your solidity test in tests/e2e/solidity/suites.go
11- Run your e2e precompile Solidity tests with 'E2E=true ./scripts/run.sh'

*/

package rewardmanager

import (
	"fmt"

	"github.com/ava-labs/subnet-evm/precompile/config"
	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/precompile/modules"

	"github.com/ethereum/go-ethereum/common"
)

var _ contract.Configurator = &configurator{}

// ConfigKey is the key used in json config files to specify this precompile config.
// must be unique across all precompiles.
const ConfigKey = "rewardManagerConfig"

// ContractAddress is the defined address of the precompile contract.
// This should be unique across all precompile contracts.
// See params/precompile_modules.go for registered precompile contracts and more information.
var ContractAddress = common.HexToAddress("{ASUITABLEHEXADDRESS}") // SET A SUITABLE HEX ADDRESS HERE

// Module is the precompile module. It is used to register the precompile contract.
var Module = modules.Module{
	ConfigKey:    ConfigKey,
	Address:      ContractAddress,
	Contract:     RewardManagerPrecompile,
	Configurator: &configurator{},
}

type configurator struct{}

func init() {
	// Register the precompile module.
	// Each precompile contract registers itself through [RegisterModule] function.
	if err := modules.RegisterModule(Module); err != nil {
		panic(err)
	}
}

// NewConfig returns a new precompile config.
// This is required for Marshal/Unmarshal the precompile config.
func (*configurator) NewConfig() config.Config {
	return &Config{}
}

// Configure configures [state] with the given [cfg] config.
// This function is called by the EVM once per precompile contract activation.
// You can use this function to set up your precompile contract's initial state,
// by using the [cfg] config and [state] stateDB.
func (*configurator) Configure(chainConfig contract.ChainConfig, cfg config.Config, state contract.StateDB, _ contract.BlockContext) error {
	config, ok := cfg.(*Config)
	if !ok {
		return fmt.Errorf("incorrect config %T: %v", config, config)
	}
	// CUSTOM CODE STARTS HERE

	// AllowList is activated for this precompile. Configuring allowlist addresses here.
	return config.Configure(state, ContractAddress)

}