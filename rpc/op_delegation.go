// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure Delegation implements the TypedOperation interface.
var _ TypedOperation = (*Delegation)(nil)

// Delegation represents a transaction operation
type Delegation struct {
	Manager
	Delegate mavryk.Address `json:"delegate,omitempty"`
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (d Delegation) Addresses(set *mavryk.AddressSet) {
	set.AddUnique(d.Source)
	if d.Delegate.IsValid() {
		set.AddUnique(d.Delegate)
	}
}
