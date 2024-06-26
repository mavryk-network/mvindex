// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

// Ensure DAL types implement the TypedOperation interface.
var (
	_ TypedOperation = (*DalPublishSlotHeader)(nil)
	_ TypedOperation = (*DalAttestation)(nil)
)

type DalPublishSlotHeader struct {
	Manager
	SlotHeader struct {
		Level      int64           `json:"level"`
		Index      byte            `json:"index"`
		Commitment string          `json:"commitment"`
		Proof      mavryk.HexBytes `json:"commitment_proof"`
	} `json:"slot_header"`
}

type DalAttestation struct {
	Generic
	Attestor    mavryk.Address `json:"attestor"`
	Attestation mavryk.Z       `json:"attestation"`
	Level       int64          `json:"level"`
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (t DalAttestation) Addresses(set *mavryk.AddressSet) {
	set.AddUnique(t.Attestor)
}
