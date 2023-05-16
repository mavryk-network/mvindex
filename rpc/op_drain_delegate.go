// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// Ensure DrainDelegate implements the TypedOperation interface.
var _ TypedOperation = (*DrainDelegate)(nil)

// DrainDelegate represents a transaction operation
type DrainDelegate struct {
	Generic
	ConsensusKey tezos.Address `json:"consensus_key"`
	Delegate     tezos.Address `json:"delegate"`
	Destination  tezos.Address `json:"destination"`
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (t DrainDelegate) Addresses(set *tezos.AddressSet) {
	set.AddUnique(t.ConsensusKey)
	set.AddUnique(t.Delegate)
	set.AddUnique(t.Destination)
}
