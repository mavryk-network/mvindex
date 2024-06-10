// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure UpdateConsensusKey implements the TypedOperation interface.
var _ TypedOperation = (*UpdateConsensusKey)(nil)

// UpdateConsensusKey represents a transaction operation
type UpdateConsensusKey struct {
	Manager
	Pk mavryk.Key `json:"pk"`
}

// Costs returns operation cost to implement TypedOperation interface.
func (t UpdateConsensusKey) Costs() mavryk.Costs {
	return mavryk.Costs{
		Fee:     t.Manager.Fee,
		GasUsed: t.Metadata.Result.Gas(),
	}
}
