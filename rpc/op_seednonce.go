// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure SeedNonce implements the TypedOperation interface.
var _ TypedOperation = (*SeedNonce)(nil)

// SeedNonce represents a seed_nonce_revelation operation
type SeedNonce struct {
	Generic
	Level int64           `json:"level"`
	Nonce mavryk.HexBytes `json:"nonce"`
}

// Ensure VdfRevelation implements the TypedOperation interface.
var _ TypedOperation = (*VdfRevelation)(nil)

// VdfRevelation represents a vdf_revelation operation
type VdfRevelation struct {
	Generic
	Solution []mavryk.HexBytes `json:"solution"`
}
