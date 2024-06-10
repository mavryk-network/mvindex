// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure Activation implements the TypedOperation interface.
var _ TypedOperation = (*Activation)(nil)

// Activation represents a transaction operation
type Activation struct {
	Generic
	Pkh    mavryk.Address  `json:"pkh"`
	Secret mavryk.HexBytes `json:"secret"`
}

// Fees returns fee-related balance updates to implement TypedOperation interface.
func (a Activation) Fees() BalanceUpdates {
	return a.Metadata.BalanceUpdates
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (a Activation) Addresses(set *mavryk.AddressSet) {
	set.AddUnique(a.Pkh)
}
