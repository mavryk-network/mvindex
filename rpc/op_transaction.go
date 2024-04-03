// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/micheline"
)

// Ensure Transaction implements the TypedOperation interface.
var _ TypedOperation = (*Transaction)(nil)

// Transaction represents a transaction operation
type Transaction struct {
	Manager
	Destination mavryk.Address       `json:"destination"`
	Amount      int64                `json:"amount,string"`
	Parameters  micheline.Parameters `json:"parameters"`
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (t Transaction) Addresses(set *mavryk.AddressSet) {
	set.AddUnique(t.Source)
	set.AddUnique(t.Destination)
	for _, v := range t.Meta().InternalResults {
		set.AddUnique(v.Source)
		set.AddUnique(v.Destination)
	}
}

func (t Transaction) AddEmbeddedAddresses(addUnique func(mavryk.Address)) {
	if !t.Destination.IsContract() {
		return
	}
	collect := func(p micheline.Prim) error {
		switch {
		case len(p.String) == 36 || len(p.String) == 37:
			if a, err := mavryk.ParseAddress(p.String); err == nil {
				addUnique(a)
			}
			return micheline.PrimSkip
		case mavryk.IsAddressBytes(p.Bytes):
			a := mavryk.Address{}
			if err := a.Decode(p.Bytes); err == nil {
				addUnique(a)
			}
			return micheline.PrimSkip
		default:
			return nil
		}
	}

	// from params
	_ = t.Parameters.Value.Walk(collect)

	// from storage
	_ = t.Metadata.Result.Storage.Walk(collect)

	// from bigmap updates
	for _, v := range t.Metadata.Result.BigmapEvents() {
		if v.Action != micheline.DiffActionUpdate && v.Action != micheline.DiffActionRemove {
			continue
		}
		vp := v.Key
		if vp.IsPacked() {
			vp, _ = vp.Unpack()
		}
		_ = vp.Walk(collect)
		vv := v.Value
		if vv.IsPacked() {
			vv, _ = vv.Unpack()
		}
		_ = vv.Walk(collect)
	}

	// from ticket updates
	for _, it := range t.Metadata.Result.TicketUpdates() {
		for _, v := range it.Updates {
			addUnique(v.Account)
		}
	}

	// from internal results
	for _, it := range t.Metadata.InternalResults {
		// from params
		_ = it.Parameters.Value.Walk(collect)

		// from origination storage
		if it.Script != nil {
			_ = it.Script.Storage.Walk(collect)
		}

		// from result storage
		_ = it.Result.Storage.Walk(collect)

		// from bigmap updates
		for _, v := range it.Result.BigmapEvents() {
			if v.Action != micheline.DiffActionUpdate && v.Action != micheline.DiffActionRemove {
				continue
			}
			vp := v.Key
			if vp.IsPacked() {
				vp, _ = vp.Unpack()
			}
			_ = vp.Walk(collect)
			vv := v.Value
			if vv.IsPacked() {
				vv, _ = vv.Unpack()
			}
			_ = vv.Walk(collect)
		}

		// from ticket updates
		for _, v := range it.Result.TicketUpdates() {
			for _, vv := range v.Updates {
				addUnique(vv.Account)
			}
		}
	}
}

type InternalResult struct {
	Kind        mavryk.OpType        `json:"kind"`
	Source      mavryk.Address       `json:"source"`
	Nonce       int64                `json:"nonce"`
	Result      OperationResult      `json:"result"`
	Destination mavryk.Address       `json:"destination"`    // transaction
	Delegate    mavryk.Address       `json:"delegate"`       // delegation
	Parameters  micheline.Parameters `json:"parameters"`     // transaction
	Amount      int64                `json:"amount,string"`  // transaction
	Balance     int64                `json:"balance,string"` // origination
	Script      *micheline.Script    `json:"script"`         // origination
	Type        micheline.Prim       `json:"type"`           // event
	Payload     micheline.Prim       `json:"payload"`        // event
	Tag         string               `json:"tag"`            // event
}

// found in block metadata from v010+
type ImplicitResult struct {
	Kind                mavryk.OpType     `json:"kind"`
	BalanceUpdates      BalanceUpdates    `json:"balance_updates"`
	ConsumedGas         int64             `json:"consumed_gas,string"`
	ConsumedMilliGas    int64             `json:"consumed_milligas,string"`
	Storage             micheline.Prim    `json:"storage"`
	StorageSize         int64             `json:"storage_size,string"`
	OriginatedContracts []mavryk.Address  `json:"originated_contracts,omitempty"`
	PaidStorageSizeDiff int64             `json:"paid_storage_size_diff,string"`
	Script              *micheline.Script `json:"script"`
}

func (r ImplicitResult) Gas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas / 1000
	}
	return r.ConsumedGas
}

func (r ImplicitResult) MilliGas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas
	}
	return r.ConsumedGas * 1000
}
