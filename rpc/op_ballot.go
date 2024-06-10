// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure Ballot implements the TypedOperation interface.
var _ TypedOperation = (*Ballot)(nil)

// Ballot represents a ballot operation
type Ballot struct {
	Generic
	Source   mavryk.Address      `json:"source"`
	Period   int                 `json:"period"`
	Ballot   mavryk.BallotVote   `json:"ballot"` // yay, nay, pass
	Proposal mavryk.ProtocolHash `json:"proposal"`
}
