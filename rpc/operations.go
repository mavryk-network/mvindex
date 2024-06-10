// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/micheline"
)

// Operation represents a single operation or batch of operations included in a block
type Operation struct {
	Hash     mavryk.OpHash    `json:"hash"`
	Contents OperationList    `json:"contents"`
	Errors   []OperationError `json:"error,omitempty"`    // mempool only
	Metadata string           `json:"metadata,omitempty"` // contains `too large` when stripped, this is BAD!!
}

// Addresses lists all Tezos addresses that appear in this operation group. This does
// not include addresses used in contract call parameters, storage updates and tickets.
func (o Operation) Addresses() *mavryk.AddressSet {
	set := mavryk.NewAddressSet()
	for _, v := range o.Contents {
		v.Addresses(set)
	}
	return set
}

func (o Operation) IsSuccess() bool {
	if len(o.Contents) == 0 {
		return false
	}
	return o.Contents[0].Result().IsSuccess()
}

// TypedOperation must be implemented by all operations
type TypedOperation interface {
	Kind() mavryk.OpType
	Meta() OperationMetadata
	Result() OperationResult
	Fees() BalanceUpdates
	Addresses(*mavryk.AddressSet)
}

// OperationError represents data describing an error conditon that lead to a
// failed operation execution.
type OperationError struct {
	NodeError
	// whitelist commonly useful error contents, avoid storing large scripts, etc
	Contract       string          `json:"contract,omitempty"`
	ContractHandle string          `json:"contract_handle,omitempty"`
	BigMap         int64           `json:"big_map,omitempty"`
	Identifier     string          `json:"identifier,omitempty"`
	Location       int64           `json:"location,omitempty"`
	Loc            int64           `json:"loc,omitempty"`
	With           *micheline.Prim `json:"with,omitempty"`
	Amount         string          `json:"amount,omitempty"`
	Balance        string          `json:"balance,omitempty"`
}

// OperationMetadata contains execution receipts for successful and failed
// operations.
type OperationMetadata struct {
	BalanceUpdates BalanceUpdates  `json:"balance_updates,omitempty"` // fee-related
	Result         OperationResult `json:"operation_result"`

	// transaction only
	InternalResults []*InternalResult `json:"internal_operation_results,omitempty"`

	// endorsement only
	Delegate            mavryk.Address `json:"delegate"`
	Slots               []int          `json:"slots,omitempty"`                // < v12
	EndorsementPower    int            `json:"endorsement_power,omitempty"`    // v12+
	PreendorsementPower int            `json:"preendorsement_power,omitempty"` // v12+
	ConsensusPower      int            `json:"consensus_power,omitempty"`      // v18+
}

func (m OperationMetadata) Power() int {
	// only one of these fields is used per operation depending on protocol
	return m.ConsensusPower + // v18+
		m.EndorsementPower + // v12+
		m.PreendorsementPower + // v12+
		len(m.Slots) // v0+
}

// Address returns the delegate address for endorsements.
func (m OperationMetadata) Address() mavryk.Address {
	return m.Delegate
}

func (m OperationMetadata) Balances() BalanceUpdates {
	return m.BalanceUpdates
}

// OperationResult contains receipts for executed operations, both success and failed.
// This type is a generic container for all possible results. Which fields are actually
// used depends on operation type and performed actions.
type OperationResult struct {
	Status               mavryk.OpStatus  `json:"status"`
	BalanceUpdates       BalanceUpdates   `json:"balance_updates"` // burn, etc
	ConsumedGas          int64            `json:"consumed_gas,string"`
	ConsumedMilliGas     int64            `json:"consumed_milligas,string"` // v007+
	Errors               []OperationError `json:"errors,omitempty"`
	Allocated            bool             `json:"allocated_destination_contract"` // tx only
	Storage              micheline.Prim   `json:"storage,omitempty"`              // tx, orig
	OriginatedContracts  []mavryk.Address `json:"originated_contracts"`           // orig only
	StorageSize          int64            `json:"storage_size,string"`            // tx, orig, const
	PaidStorageSizeDiff  int64            `json:"paid_storage_size_diff,string"`  // tx, orig
	BigmapDiff           json.RawMessage  `json:"big_map_diff,omitempty"`         // tx, orig, <v013
	LazyStorageDiff      json.RawMessage  `json:"lazy_storage_diff,omitempty"`    // v008+ tx, orig
	GlobalAddress        mavryk.ExprHash  `json:"global_address"`                 // global constant
	TicketUpdatesCorrect []TicketUpdate   `json:"ticket_updates"`                 // v015, correct name on external
	TicketReceipts       []TicketUpdate   `json:"ticket_receipt"`                 // v015, incorrect name on internal

	// v013 tx rollup
	TxRollupResult

	// v016 smart rollup
	SmartRollupResult

	// internal
	bigmapEvents micheline.BigmapEvents
}

// Always use this helper to retrieve Ticket updates. This is because due to
// lack of quality control Tezos Lima protocol ended up with 2 distinct names
// for ticket updates in external call receipts versus internal call receipts.
func (r OperationResult) TicketUpdates() []TicketUpdate {
	if len(r.TicketUpdatesCorrect) > 0 {
		return r.TicketUpdatesCorrect
	}
	return r.TicketReceipts
}

func (r OperationResult) BigmapEvents() micheline.BigmapEvents {
	switch {
	case r.bigmapEvents != nil:
		// skip
	case r.LazyStorageDiff != nil:
		res := make(micheline.LazyEvents, 0)
		_ = json.Unmarshal(r.LazyStorageDiff, &res)
		r.bigmapEvents = res.BigmapEvents()
	case r.BigmapDiff != nil:
		r.bigmapEvents = make(micheline.BigmapEvents, 0)
		_ = json.Unmarshal(r.BigmapDiff, &r.bigmapEvents)
	}
	return r.bigmapEvents
}

func (r OperationResult) Balances() BalanceUpdates {
	return r.BalanceUpdates
}

func (r OperationResult) IsSuccess() bool {
	return r.Status == mavryk.OpStatusApplied
}

func (r OperationResult) Gas() int64 {
	if r.ConsumedMilliGas > 0 {
		var corr int64
		if r.ConsumedMilliGas%1000 > 0 {
			corr++
		}
		return r.ConsumedMilliGas/1000 + corr
	}
	return r.ConsumedGas
}

func (r OperationResult) MilliGas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas
	}
	return r.ConsumedGas * 1000
}

// Generic is the most generic operation type.
type Generic struct {
	OpKind   mavryk.OpType      `json:"kind"`
	Metadata *OperationMetadata `json:"metadata,omitempty"`
}

// Kind returns the operation's type. Implements TypedOperation interface.
func (e Generic) Kind() mavryk.OpType {
	return e.OpKind
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (e Generic) Meta() OperationMetadata {
	return *e.Metadata
}

// Result returns an empty operation result to implement TypedOperation interface.
func (e Generic) Result() OperationResult {
	return e.Metadata.Result
}

// Fees returns an empty balance update list to implement TypedOperation interface.
func (e Generic) Fees() BalanceUpdates {
	return e.Metadata.BalanceUpdates
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (e Generic) Addresses(set *mavryk.AddressSet) {
	// noop
}

// Manager represents data common for all manager operations.
type Manager struct {
	Generic
	Source       mavryk.Address `json:"source"`
	Fee          int64          `json:"fee,string"`
	Counter      int64          `json:"counter,string"`
	GasLimit     int64          `json:"gas_limit,string"`
	StorageLimit int64          `json:"storage_limit,string"`
}

// Limits returns manager operation limits to implement TypedOperation interface.
func (e Manager) Limits() mavryk.Limits {
	return mavryk.Limits{
		Fee:          e.Fee,
		GasLimit:     e.GasLimit,
		StorageLimit: e.StorageLimit,
	}
}

// Addresses adds all addresses used in this operation to the set.
// Implements TypedOperation interface.
func (e Manager) Addresses(set *mavryk.AddressSet) {
	set.AddUnique(e.Source)
}

// OperationList is a slice of TypedOperation (interface type) with custom JSON unmarshaller
type OperationList []TypedOperation

// UnmarshalJSON implements json.Unmarshaler
func (e *OperationList) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 {
		return nil
	}

	if data[0] != '[' {
		return fmt.Errorf("rpc: expected operation array")
	}

	// fmt.Printf("Decoding ops: %s\n", string(data))
	dec := json.NewDecoder(bytes.NewReader(data))

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return fmt.Errorf("rpc: %v", err)
	}

	for dec.More() {
		// peek into `{"kind":"...",` field
		start := int(dec.InputOffset()) + 9
		// after first JSON object, decoder pos is at `,`
		if data[start] == '"' {
			start += 1
		}
		end := start + bytes.IndexByte(data[start:], '"')
		kind := mavryk.ParseOpType(string(data[start:end]))
		var op TypedOperation
		switch kind {
		// anonymous operations
		case mavryk.OpTypeActivateAccount:
			op = &Activation{}
		case mavryk.OpTypeDoubleBakingEvidence:
			op = &DoubleBaking{}
		case mavryk.OpTypeDoubleEndorsementEvidence,
			mavryk.OpTypeDoublePreendorsementEvidence:
			op = &DoubleEndorsement{}
		case mavryk.OpTypeSeedNonceRevelation:
			op = &SeedNonce{}
		case mavryk.OpTypeDrainDelegate:
			op = &DrainDelegate{}

		// consensus operations
		case mavryk.OpTypeEndorsement,
			mavryk.OpTypeEndorsementWithSlot,
			mavryk.OpTypePreendorsement:
			op = &Endorsement{}

		// amendment operations
		case mavryk.OpTypeProposals:
			op = &Proposals{}
		case mavryk.OpTypeBallot:
			op = &Ballot{}

		// manager operations
		case mavryk.OpTypeTransaction:
			op = &Transaction{}
		case mavryk.OpTypeOrigination:
			op = &Origination{}
		case mavryk.OpTypeDelegation:
			op = &Delegation{}
		case mavryk.OpTypeReveal:
			op = &Reveal{}
		case mavryk.OpTypeRegisterConstant:
			op = &ConstantRegistration{}
		case mavryk.OpTypeSetDepositsLimit:
			op = &SetDepositsLimit{}
		case mavryk.OpTypeIncreasePaidStorage:
			op = &IncreasePaidStorage{}
		case mavryk.OpTypeVdfRevelation:
			op = &VdfRevelation{}
		case mavryk.OpTypeTransferTicket:
			op = &TransferTicket{}
		case mavryk.OpTypeUpdateConsensusKey:
			op = &UpdateConsensusKey{}

			// DEPRECATED: tx rollup operations, kept for testnet backward compatibility
		case mavryk.OpTypeTxRollupOrigination,
			mavryk.OpTypeTxRollupSubmitBatch,
			mavryk.OpTypeTxRollupCommit,
			mavryk.OpTypeTxRollupReturnBond,
			mavryk.OpTypeTxRollupFinalizeCommitment,
			mavryk.OpTypeTxRollupRemoveCommitment,
			mavryk.OpTypeTxRollupRejection,
			mavryk.OpTypeTxRollupDispatchTickets:
			op = &TxRollup{}

		case mavryk.OpTypeSmartRollupOriginate:
			op = &SmartRollupOriginate{}
		case mavryk.OpTypeSmartRollupAddMessages:
			op = &SmartRollupAddMessages{}
		case mavryk.OpTypeSmartRollupCement:
			op = &SmartRollupCement{}
		case mavryk.OpTypeSmartRollupPublish:
			op = &SmartRollupPublish{}
		case mavryk.OpTypeSmartRollupRefute:
			op = &SmartRollupRefute{}
		case mavryk.OpTypeSmartRollupTimeout:
			op = &SmartRollupTimeout{}
		case mavryk.OpTypeSmartRollupExecuteOutboxMessage:
			op = &SmartRollupExecuteOutboxMessage{}
		case mavryk.OpTypeSmartRollupRecoverBond:
			op = &SmartRollupRecoverBond{}
		case mavryk.OpTypeDalAttestation:
			op = &DalAttestation{}
		case mavryk.OpTypeDalPublishSlotHeader:
			op = &DalPublishSlotHeader{}

		default:
			return fmt.Errorf("rpc: unsupported op %q", string(data[start:end]))
		}

		if err := dec.Decode(op); err != nil {
			return fmt.Errorf("rpc: operation kind %s: %w", kind, err)
		}
		(*e) = append(*e, op)
	}

	return nil
}
