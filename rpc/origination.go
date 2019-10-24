// Copyright (c) 2019 KIDTSUNAMI
// Author: alex@kidtsunami.com

package rpc

import (
	"blockwatch.cc/tzindex/chain"
	"blockwatch.cc/tzindex/micheline"
)

// OriginationOp represents a contract creation operation
type OriginationOp struct {
	GenericOp
	Source         chain.Address          `json:"source"`
	Fee            int64                  `json:"fee,string"`
	Counter        int64                  `json:"counter,string"`
	GasLimit       int64                  `json:"gas_limit,string"`
	StorageLimit   int64                  `json:"storage_limit,string"`
	ManagerPubkey  chain.Address          `json:"manager_pubkey"` // proto v1 & >= v4
	ManagerPubkey2 chain.Address          `json:"managerPubkey"`  // proto v2, v3
	Balance        int64                  `json:"balance,string"`
	Spendable      bool                   `json:"spendable"`
	Delegatable    bool                   `json:"delegatable"`
	Delegate       *chain.Address         `json:"delegate"`
	Script         *micheline.Script      `json:"script"`
	Metadata       *OriginationOpMetadata `json:"metadata"`
}

// OriginationOpMetadata represents a transaction operation metadata
type OriginationOpMetadata struct {
	BalanceUpdates BalanceUpdates    `json:"balance_updates"` // fee-related
	Result         OriginationResult `json:"operation_result"`
}

// OriginationResult represents a contract creation result
type OriginationResult struct {
	BalanceUpdates      BalanceUpdates   `json:"balance_updates"` // burned fees
	OriginatedContracts []chain.Address  `json:"originated_contracts"`
	ConsumedGas         int64            `json:"consumed_gas,string"`
	StorageSize         int64            `json:"storage_size,string"`
	PaidStorageSizeDiff int64            `json:"paid_storage_size_diff,string"`
	Status              chain.OpStatus   `json:"status"`
	Errors              []OperationError `json:"errors,omitempty"`
}