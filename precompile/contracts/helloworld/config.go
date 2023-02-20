// Code generated
// This file is a generated precompile contract config with stubbed abstract functions.
// The file is generated by a template. Please inspect every code and comment in this file before use.

// There are some must-be-done changes waiting in the file. Each area requiring you to add your code is marked with CUSTOM CODE to make them easy to find and modify.
// Additionally there are other files you need to edit to activate your precompile.
// These areas are highlighted with comments "ADD YOUR PRECOMPILE HERE".
// For testing take a look at other precompile tests in contract_test.go and config_test.go in other precompile folders.
// See the tutorial in https://docs.avax.network/subnets/hello-world-precompile-tutorial for more information about precompile development.

/* General guidelines for precompile development:
1- Set a suitable config key in generated module.go. E.g: "yourPrecompileConfig"
2- Read the comment and set a suitable contract address in generated module.go. E.g:
	ContractAddress = common.HexToAddress("ASUITABLEHEXADDRESS")
3- It is recommended to only modify code in the highlighted areas marked with "CUSTOM CODE STARTS HERE". Typically, custom codes are required in only those areas.
Modifying code outside of these areas should be done with caution and with a deep understanding of how these changes may impact the EVM.
4- Set gas costs in generated contract.go
5- Add your config unit tests under generated package config_test.go
6- Add your contract unit tests undertgenerated package contract_test.go
7- Additionally you can add a full-fledged VM test for your precompile under plugin/vm/vm_test.go. See existing precompile tests for examples.
8- Add your solidity interface and test contract to contract-examples/contracts
9- Write solidity tests for your precompile in contract-examples/test
10- Create your genesis with your precompile enabled in tests/e2e/genesis/
11- Create e2e test for your solidity test in tests/e2e/solidity/suites.go
12- Run your e2e precompile Solidity tests with 'E2E=true ./scripts/run.sh'
*/

package helloworld

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/allowlist"
	"github.com/ava-labs/subnet-evm/precompile/precompileconfig"

	"github.com/ethereum/go-ethereum/common"
)

var _ precompileconfig.Config = &Config{}

// Config implements the StatefulPrecompileConfig
// interface while adding in the HelloWorld specific precompile address.
type Config struct {
	allowlist.AllowListConfig
	precompileconfig.Upgrade
}

// NewConfig returns a config for a network upgrade at [blockTimestamp] that enables
// HelloWorld with the given [admins] and [enableds] as members of the allowlist.
// HelloWorld  with the given [admins] as members of the allowlist .
func NewConfig(blockTimestamp *big.Int, admins []common.Address, enableds []common.Address) *Config {
	return &Config{
		AllowListConfig: allowlist.AllowListConfig{
			AdminAddresses:   admins,
			EnabledAddresses: enableds,
		},
		Upgrade: precompileconfig.Upgrade{BlockTimestamp: blockTimestamp},
	}
}

// NewDisableConfig returns config for a network upgrade at [blockTimestamp]
// that disables HelloWorld.
func NewDisableConfig(blockTimestamp *big.Int) *Config {
	return &Config{
		Upgrade: precompileconfig.Upgrade{
			BlockTimestamp: blockTimestamp,
			Disable:        true,
		},
	}
}

// Key returns the key for the HelloWorld precompileconfig.
// This should be the same key as used in the precompile module.
func (*Config) Key() string { return ConfigKey }

// Verify tries to verify Config and returns an error accordingly.
func (c *Config) Verify() error {
	// Verify AllowList first
	if err := c.AllowListConfig.Verify(); err != nil {
		return err
	}

	// CUSTOM CODE STARTS HERE
	// Add your own custom verify code for Config here
	// and return an error accordingly
	return nil
}

// Equal returns true if [s] is a [*Config] and it has been configured identical to [c].
func (c *Config) Equal(s precompileconfig.Config) bool {
	// typecast before comparison
	other, ok := (s).(*Config)
	if !ok {
		return false
	}
	// CUSTOM CODE STARTS HERE
	// modify this boolean accordingly with your custom Config, to check if [other] and the current [c] are equal
	// if Config contains only Upgrade  and AllowListConfig  you can skip modifying it.
	equals := c.Upgrade.Equal(&other.Upgrade) && c.AllowListConfig.Equal(&other.AllowListConfig)
	return equals
}
