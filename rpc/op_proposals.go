// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure Proposals implements the TypedOperation interface.
var _ TypedOperation = (*Proposals)(nil)

// Proposals represents a proposal operation
type Proposals struct {
	Generic
	Source    mavryk.Address        `json:"source"`
	Period    int                   `json:"period"`
	Proposals []mavryk.ProtocolHash `json:"proposals"`
}
