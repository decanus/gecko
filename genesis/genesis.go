// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

// TODO: Move this to a separate repo and leave only a byte array

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ava-labs/gecko/ids"
	"github.com/ava-labs/gecko/vms/avm"
	"github.com/ava-labs/gecko/vms/components/codec"
	"github.com/ava-labs/gecko/vms/evm"
	"github.com/ava-labs/gecko/vms/platformvm"
	"github.com/ava-labs/gecko/vms/spchainvm"
	"github.com/ava-labs/gecko/vms/spdagvm"
	"github.com/ava-labs/gecko/vms/timestampvm"
	"github.com/ethereum/go-ethereum/core"
)

// Note that since an AVA network has exactly one Platform Chain,
// and the Platform Chain defines the genesis state of the network
// (who is staking, which chains exist, etc.), defining the genesis
// state of the Platform Chain is the same as defining the genesis
// state of the network.

// Hardcoded network IDs
const (
	MainnetID  uint32 = 1
	TestnetID  uint32 = 2
	BorealisID uint32 = 2
	LocalID    uint32 = 12345

	MainnetName  = "mainnet"
	TestnetName  = "testnet"
	BorealisName = "borealis"
	LocalName    = "local"
)

var (
	validNetworkName = regexp.MustCompile(`network-[0-9]+`)
)

// Hard coded genesis constants
var (
	// Give special names to the mainnet and testnet
	NetworkIDToNetworkName = map[uint32]string{
		MainnetID: MainnetName,
		TestnetID: BorealisName,
		LocalID:   LocalName,
	}
	NetworkNameToNetworkID = map[string]uint32{
		MainnetName:  MainnetID,
		TestnetName:  TestnetID,
		BorealisName: BorealisID,
		LocalName:    LocalID,
	}
	Keys = []string{
		"ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN",
	}
	Addresses = []string{
		"6Y3kysjF9jnHnYkdS9yGAuoHyae2eNmeV",
	}
	ParsedAddresses = []ids.ShortID{}
	StakerIDs       = []string{
		"7Xhw2mDxuDS44j42TCB6U5579esbSt3Lg",
		"MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ",
		"NFBbbJ4qCmNaCzeW7sxErhvWqvEQMnYcN",
		"GWPcbFJZFfZreETSoWjPimr846mXEKCtu",
		"P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
	}
	ParsedStakerIDs = []ids.ShortID{}
)

func init() {
	for _, addrStr := range Addresses {
		addr, err := ids.ShortFromString(addrStr)
		if err != nil {
			panic(err)
		}
		ParsedAddresses = append(ParsedAddresses, addr)
	}
	for _, stakerIDStr := range StakerIDs {
		stakerID, err := ids.ShortFromString(stakerIDStr)
		if err != nil {
			panic(err)
		}
		ParsedStakerIDs = append(ParsedStakerIDs, stakerID)
	}
}

// NetworkName returns a human readable name for the network with
// ID [networkID]
func NetworkName(networkID uint32) string {
	if name, exists := NetworkIDToNetworkName[networkID]; exists {
		return name
	}
	return fmt.Sprintf("network-%d", networkID)
}

// NetworkID returns the ID of the network with name [networkName]
func NetworkID(networkName string) (uint32, error) {
	networkName = strings.ToLower(networkName)
	if id, exists := NetworkNameToNetworkID[networkName]; exists {
		return id, nil
	}

	if id, err := strconv.ParseUint(networkName, 10, 0); err == nil {
		if id > math.MaxUint32 {
			return 0, fmt.Errorf("NetworkID %s not in [0, 2^32)", networkName)
		}
		return uint32(id), nil
	}
	if validNetworkName.MatchString(networkName) {
		if id, err := strconv.Atoi(networkName[8:]); err == nil {
			if id > math.MaxUint32 {
				return 0, fmt.Errorf("NetworkID %s not in [0, 2^32)", networkName)
			}
			return uint32(id), nil
		}
	}

	return 0, fmt.Errorf("Failed to parse %s as a network name", networkName)
}

// Aliases returns the default aliases based on the network ID
func Aliases(networkID uint32) (generalAliases map[string][]string, chainAliases map[[32]byte][]string, vmAliases map[[32]byte][]string) {
	generalAliases = map[string][]string{
		"vm/" + platformvm.ID.String():  []string{"vm/platform"},
		"vm/" + avm.ID.String():         []string{"vm/avm"},
		"vm/" + evm.ID.String():         []string{"vm/evm"},
		"vm/" + spdagvm.ID.String():     []string{"vm/spdag"},
		"vm/" + spchainvm.ID.String():   []string{"vm/spchain"},
		"vm/" + timestampvm.ID.String(): []string{"vm/timestamp"},
		"bc/" + ids.Empty.String():      []string{"P", "platform", "bc/P", "bc/platform"},
	}
	chainAliases = map[[32]byte][]string{
		ids.Empty.Key(): []string{"P", "platform"},
	}
	vmAliases = map[[32]byte][]string{
		platformvm.ID.Key():  []string{"platform"},
		avm.ID.Key():         []string{"avm"},
		evm.ID.Key():         []string{"evm"},
		spdagvm.ID.Key():     []string{"spdag"},
		spchainvm.ID.Key():   []string{"spchain"},
		timestampvm.ID.Key(): []string{"timestamp"},
	}

	genesisBytes := Genesis(networkID)
	genesis := &platformvm.Genesis{}                  // TODO let's not re-create genesis to do aliasing
	platformvm.Codec.Unmarshal(genesisBytes, genesis) // TODO check for error
	genesis.Initialize()

	for _, chain := range genesis.Chains {
		switch {
		case avm.ID.Equals(chain.VMID):
			generalAliases["bc/"+chain.ID().String()] = []string{"X", "avm", "bc/X", "bc/avm"}
			chainAliases[chain.ID().Key()] = []string{"X", "avm"}
		case evm.ID.Equals(chain.VMID):
			generalAliases["bc/"+chain.ID().String()] = []string{"C", "evm", "bc/C", "bc/evm"}
			chainAliases[chain.ID().Key()] = []string{"C", "evm"}
		case spdagvm.ID.Equals(chain.VMID):
			generalAliases["bc/"+chain.ID().String()] = []string{"bc/spdag"}
			chainAliases[chain.ID().Key()] = []string{"spdag"}
		case spchainvm.ID.Equals(chain.VMID):
			generalAliases["bc/"+chain.ID().String()] = []string{"bc/spchain"}
			chainAliases[chain.ID().Key()] = []string{"spchain"}
		case timestampvm.ID.Equals(chain.VMID):
			generalAliases["bc/"+chain.ID().String()] = []string{"bc/timestamp"}
			chainAliases[chain.ID().Key()] = []string{"timestamp"}
		}
	}
	return
}

// Genesis returns the genesis data of the Platform Chain.
// Since the Platform Chain causes the creation of all other
// chains, this function returns the genesis data of the entire network.
// The ID of the new network is [networkID].
func Genesis(networkID uint32) []byte {
	if networkID != LocalID {
		panic("unknown network ID provided")
	}

	genesisState := platformvm.Genesis{
		Accounts:  	make([]platformvm.Account, 0),
		Validators: &platformvm.EventHeap{},
		Chains:     make([]*platformvm.CreateChainTx, 0),
		Timestamp:  0,
	}

	genesisBytes, err := platformvm.Codec.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return genesisBytes
}

// VMGenesis ...
func VMGenesis(networkID uint32, vmID ids.ID) *platformvm.CreateChainTx {
	genesisBytes := Genesis(networkID)
	genesis := platformvm.Genesis{}
	platformvm.Codec.Unmarshal(genesisBytes, &genesis)
	for _, chain := range genesis.Chains {
		if chain.VMID.Equals(vmID) {
			return chain
		}
	}
	return nil
}
