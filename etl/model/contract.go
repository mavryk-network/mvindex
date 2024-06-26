// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package model

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"blockwatch.cc/packdb/pack"
	"blockwatch.cc/packdb/util"
	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/micheline"
	"github.com/mavryk-network/mvindex/rpc"
)

const ContractTableKey = "contract"

var (
	contractPool = &sync.Pool{
		New: func() interface{} { return new(Contract) },
	}

	contractSize = int(reflect.TypeOf(Contract{}).Size())

	ErrNoContract = errors.New("contract not indexed")
)

type ContractID uint64

func (id ContractID) U64() uint64 {
	return uint64(id)
}

// Contract holds code and info about smart contracts on the Tezos blockchain.
type Contract struct {
	RowId          ContractID           `pack:"I,pk"      json:"row_id"`
	Address        mavryk.Address       `pack:"H,bloom=3" json:"address"`
	AccountId      AccountID            `pack:"A,u32"     json:"account_id"`
	CreatorId      AccountID            `pack:"C,u32"     json:"creator_id"`
	FirstSeen      int64                `pack:"f,i32"     json:"first_seen"`
	LastSeen       int64                `pack:"l,i32"     json:"last_seen"`
	StorageSize    int64                `pack:"z,i32"     json:"storage_size"`
	StoragePaid    int64                `pack:"y,i32"     json:"storage_paid"`
	StorageBurn    int64                `pack:"Y"         json:"storage_burn"`
	Script         []byte               `pack:"s,snappy"  json:"script"`
	Storage        []byte               `pack:"g,snappy"  json:"storage"`
	InterfaceHash  uint64               `pack:"i,snappy"  json:"iface_hash"`
	CodeHash       uint64               `pack:"c,snappy"  json:"code_hash"`
	StorageHash    uint64               `pack:"x,snappy"  json:"storage_hash"`
	CallStats      []byte               `pack:"S,snappy"  json:"call_stats"`
	Features       micheline.Features   `pack:"F,snappy"  json:"features"`
	Interfaces     micheline.Interfaces `pack:"n,snappy"  json:"interfaces"`
	LedgerType     TokenType            `pack:"t"         json:"ledger_type"`
	LedgerSchema   LedgerSchema         `pack:"h"         json:"ledger_schema"`
	LedgerBigmap   int64                `pack:"b,i32"     json:"ledger_bigmap"`
	LedgerMeta     MetaID               `pack:"M"         json:"metadata_id"`
	MetadataBigmap int64                `pack:"m,i32"     json:"metadata_bigmap"`

	IsDirty bool              `pack:"-" json:"-"` // indicates an update happened
	IsNew   bool              `pack:"-" json:"-"` // new contract, used during migration
	script  *micheline.Script `pack:"-" json:"-"` // cached decoded script
	params  micheline.Type    `pack:"-" json:"-"` // cached param type
	storage micheline.Type    `pack:"-" json:"-"` // cached storage type
}

func (c *Contract) HeapSize() int {
	sz := contractSize + len(c.Script) + len(c.Storage) + len(c.CallStats)
	if c.script != nil {
		sz += len(c.Script) * 2 // approx
	}
	return sz
}

// Ensure Contract implements the pack.Item interface.
var _ pack.Item = (*Contract)(nil)

// assuming the op was successful!
func NewContract(acc *Account, oop *rpc.Origination, op *Op, dict micheline.ConstantDict, p *rpc.Params) *Contract {
	c := AllocContract()
	c.Address = acc.Address.Clone()
	c.AccountId = acc.RowId
	c.CreatorId = op.SenderId
	c.FirstSeen = op.Height
	c.LastSeen = op.Height
	res := oop.Result()
	c.StorageSize = res.StorageSize
	c.StoragePaid = res.PaidStorageSizeDiff
	c.StorageBurn += c.StoragePaid * p.CostPerByte
	if oop.Script != nil {
		c.Features = oop.Script.Features()
		if c.Features.Contains(micheline.FeatureGlobalConstant) {
			oop.Script.ExpandConstants(dict)
			c.Features |= oop.Script.Features()
		}
		c.Script, _ = oop.Script.MarshalBinary()
		c.Storage, _ = oop.Script.Storage.MarshalBinary()
		c.InterfaceHash = oop.Script.InterfaceHash()
		c.CodeHash = oop.Script.CodeHash()
		c.StorageHash = oop.Script.StorageHash()
		c.Interfaces = oop.Script.Interfaces()
		ep, _ := oop.Script.Entrypoints(false)
		c.CallStats = make([]byte, 4*len(ep))
		c.LedgerSchema, c.LedgerType, c.LedgerBigmap, c.MetadataBigmap = DetectLedger(*oop.Script)
	}
	flags := oop.BabylonFlags(p.Version)
	if flags.IsSpendable() {
		c.Features |= micheline.FeatureSpendable
	}
	if flags.IsDelegatable() {
		c.Features |= micheline.FeatureDelegatable
	}

	c.IsNew = true
	c.IsDirty = true
	return c
}

func NewInternalContract(acc *Account, iop rpc.InternalResult, op *Op, dict micheline.ConstantDict, p *rpc.Params) *Contract {
	c := AllocContract()
	c.Address = acc.Address.Clone()
	c.AccountId = acc.RowId
	c.CreatorId = op.CreatorId // may be another KT1
	c.FirstSeen = op.Height
	c.LastSeen = op.Height
	res := iop.Result
	c.StorageSize = res.StorageSize
	c.StoragePaid = res.PaidStorageSizeDiff
	c.StorageBurn += c.StoragePaid * p.CostPerByte
	if iop.Script != nil {
		c.Features = iop.Script.Features()
		if c.Features.Contains(micheline.FeatureGlobalConstant) {
			iop.Script.ExpandConstants(dict)
			c.Features |= iop.Script.Features()
		}
		c.Script, _ = iop.Script.MarshalBinary()
		c.Storage, _ = iop.Script.Storage.MarshalBinary()
		c.InterfaceHash = iop.Script.InterfaceHash()
		c.CodeHash = iop.Script.CodeHash()
		c.StorageHash = iop.Script.StorageHash()
		c.Interfaces = iop.Script.Interfaces()
		ep, _ := iop.Script.Entrypoints(false)
		c.CallStats = make([]byte, 4*len(ep))
		c.LedgerSchema, c.LedgerType, c.LedgerBigmap, c.MetadataBigmap = DetectLedger(*iop.Script)
	}
	// pre-babylon did not have any internal originations
	// c.Features |= micheline.FeatureSpendable | micheline.FeatureDelegatable
	c.IsNew = true
	c.IsDirty = true
	return c
}

func NewImplicitContract(acc *Account, res rpc.ImplicitResult, op *Op, p *rpc.Params) *Contract {
	c := AllocContract()
	c.Address = acc.Address.Clone()
	c.AccountId = acc.RowId
	c.CreatorId = acc.CreatorId
	c.FirstSeen = op.Height
	c.LastSeen = op.Height
	c.StorageSize = res.StorageSize
	c.StoragePaid = res.PaidStorageSizeDiff
	c.StorageBurn += c.StoragePaid * p.CostPerByte
	if res.Script != nil {
		c.Script, _ = res.Script.MarshalBinary()
		c.Storage, _ = res.Script.Storage.MarshalBinary()
		c.InterfaceHash = res.Script.InterfaceHash()
		c.CodeHash = res.Script.CodeHash()
		c.StorageHash = res.Script.StorageHash()
		c.Features = res.Script.Features()
		c.Interfaces = res.Script.Interfaces()
		ep, _ := res.Script.Entrypoints(false)
		c.CallStats = make([]byte, 4*len(ep))
		c.LedgerSchema, c.LedgerType, c.LedgerBigmap, c.MetadataBigmap = DetectLedger(*res.Script)
	}
	c.IsNew = true
	c.IsDirty = true
	return c
}

// create manager.tz contract, used during migration only
func NewManagerTzContract(a *Account, height int64) (*Contract, error) {
	c := AllocContract()
	c.Address = a.Address.Clone()
	c.AccountId = a.RowId
	c.CreatorId = a.CreatorId
	c.FirstSeen = a.FirstSeen
	c.LastSeen = height
	script, _ := micheline.MakeManagerScript(a.Address.Encode())
	c.Script, _ = script.MarshalBinary()
	c.Storage, _ = script.Storage.MarshalBinary()
	c.InterfaceHash = script.InterfaceHash()
	c.CodeHash = script.CodeHash()
	c.StorageHash = script.StorageHash()
	c.Features = script.Features()
	c.Interfaces = script.Interfaces()
	c.StorageSize = 232           // fixed 232 bytes
	c.StoragePaid = 0             // noone paid for this
	c.CallStats = make([]byte, 8) // 2 entrypoints, 'do' (0) and 'default' (1)
	binary.BigEndian.PutUint32(c.CallStats[4:8], uint32(a.NTxIn))
	c.IsNew = true
	c.IsDirty = true
	return c, nil
}

func NewRollupContract(acc *Account, op *Op, res rpc.OperationResult, p *rpc.Params) *Contract {
	c := AllocContract()
	c.Address = acc.Address.Clone()
	c.AccountId = acc.RowId
	c.CreatorId = op.SenderId
	c.FirstSeen = op.Height
	c.LastSeen = op.Height
	c.StorageSize = 4000 // toru fixed, tx_rollup_origination_size
	if res.Size != nil {
		c.StorageSize = res.Size.Int64()
	}
	c.StoragePaid = c.StorageSize
	c.StorageBurn += c.StoragePaid * p.CostPerByte
	// tx_rollup no script
	// 7 ops excl origination
	// +1 first is the fake `deposit` op for tickets
	//
	// smart_rollup
	// TODO: script (wrap pvm kind, kernal, params type, proof)
	// - 7 ops excl origination
	c.CallStats = make([]byte, 4*8)
	c.IsNew = true
	c.IsDirty = true
	return c
}

func AllocContract() *Contract {
	return contractPool.Get().(*Contract)
}

func (c *Contract) Free() {
	c.Reset()
	contractPool.Put(c)
}

func (c Contract) ID() uint64 {
	return uint64(c.RowId)
}

func (c *Contract) SetID(id uint64) {
	c.RowId = ContractID(id)
}

func (m Contract) TableKey() string {
	return ContractTableKey
}

func (m Contract) TableOpts() pack.Options {
	return pack.Options{
		PackSizeLog2:    15,
		JournalSizeLog2: 15,
		CacheSize:       256,
		FillLevel:       80,
	}
}

func (m Contract) IndexOpts(key string) pack.Options {
	return pack.NoOptions
}

func (c Contract) String() string {
	return c.Address.String()
}

func (c *Contract) Reset() {
	*c = Contract{}
}

func (c *Contract) Update(op *Op, p *rpc.Params) bool {
	c.LastSeen = op.Height
	c.IncCallStats(op.Entrypoint)
	c.IsDirty = true
	if op.Storage != nil && op.StorageHash != c.StorageHash {
		c.Storage = op.Storage
		c.StorageHash = op.StorageHash
		c.StorageSize = int64(len(c.Storage))
		c.StoragePaid += op.StoragePaid
		c.StorageBurn += op.StoragePaid * p.CostPerByte
		return true
	}
	return false
}

func (c *Contract) Rollback(drop, last *Op, p *rpc.Params) {
	if last != nil {
		c.LastSeen = last.Height
		if last.Storage != nil {
			c.Storage = last.Storage
			c.StorageSize = int64(len(c.Storage))
			c.StorageHash = last.StorageHash
		}
	} else if c.script != nil {
		// back to origination
		c.Storage, _ = c.script.Storage.MarshalBinary()
		c.StorageHash = c.script.Storage.Hash64()
		c.LastSeen = c.FirstSeen
		c.StorageSize = int64(len(c.Storage))
	}
	c.StoragePaid -= drop.StoragePaid
	c.StorageBurn -= drop.StoragePaid * p.CostPerByte
	c.DecCallStats(drop.Entrypoint)
	c.IsDirty = true
}

func (c Contract) IsRollup() bool {
	switch c.Address.Type() {
	case mavryk.AddressTypeTxRollup, mavryk.AddressTypeSmartRollup:
		return true
	default:
		return false
	}
}

func (c *Contract) ListTxRollupCallStats() map[string]int {
	res := make(map[string]int, len(c.CallStats)>>2)
	for i, v := range []string{
		"deposit", // fake, /sigh/ :(
		"tx_rollup_submit_batch",
		"tx_rollup_commit",
		"tx_rollup_return_bond",
		"tx_rollup_finalize_commitment",
		"tx_rollup_remove_commitment",
		"tx_rollup_rejection",
		"tx_rollup_dispatch_tickets",
	} {
		res[v] = int(binary.BigEndian.Uint32(c.CallStats[i*4:]))
	}
	return res
}

func (c *Contract) ListSmartRollupCallStats() map[string]int {
	res := make(map[string]int, len(c.CallStats)>>2)
	for i, v := range []string{
		"deposit", // fake, /sigh/ :(
		"smart_rollup_add_messages",
		"smart_rollup_cement",
		"smart_rollup_publish",
		"smart_rollup_refute",
		"smart_rollup_timeout",
		"smart_rollup_execute_outbox_message",
		"smart_rollup_recover_bond",
	} {
		res[v] = int(binary.BigEndian.Uint32(c.CallStats[i*4:]))
	}
	return res
}

func (c *Contract) ListCallStats() map[string]int {
	switch c.Address.Type() {
	case mavryk.AddressTypeTxRollup:
		return c.ListTxRollupCallStats()
	case mavryk.AddressTypeSmartRollup:
		return c.ListSmartRollupCallStats()
	}
	// list entrypoint names first
	pTyp, _, err := c.LoadType()
	if err != nil {
		return nil
	}

	ep, err := pTyp.Entrypoints(false)
	if err != nil {
		return nil
	}

	// sort entrypoint map by id, we only need names here
	byId := make([]string, len(ep))
	for _, v := range ep {
		byId[v.Id] = v.Name
	}

	res := make(map[string]int, len(c.CallStats)>>2)
	for i, name := range byId {
		res[name] = int(binary.BigEndian.Uint32(c.CallStats[i*4:]))
	}
	return res
}

func (c *Contract) NamedBigmaps(m []*BigmapAlloc) map[string]int64 {
	if len(m) == 0 {
		return nil
	}
	_, sTyp, err := c.LoadType()
	if err != nil {
		return nil
	}
	named := make(map[string]int64)

	// find bigmap typedefs in script
	bigmaps, _ := sTyp.FindOpCodes(micheline.T_BIG_MAP)

	// unpack micheline types into tzgo types for matching
	// this resolves ambiguities from different comb pair expressions
	types := make([]micheline.Typedef, len(bigmaps))
	for i, v := range bigmaps {
		types[i] = micheline.Type{Prim: v}.Typedef("")
	}

	// match bigmap allocs to type annotations using type comparison
	for i, v := range m {
		kt, vt := v.GetKeyType().Typedef(""), v.GetValueType().Typedef("")
		var name string
		for _, typ := range types {
			if !typ.Left().Unfold().Equal(kt) {
				continue
			}
			if !typ.Right().Unfold().Equal(vt) {
				continue
			}
			name = typ.Name
			// some bigmap types may be reused (different bigmaps have the same type)
			// so be careful not overwriting existing matches
			if _, ok := named[name]; !ok {
				break
			}
		}
		// generate a unique name when annots are missing
		if name == "" {
			name = "bigmap_" + strconv.Itoa(i)
		}
		// make sure name is not a duplicate
		if _, ok := named[name]; ok {
			var c int
			for {
				n := name + "_" + strconv.Itoa(c)
				if _, ok := named[n]; !ok {
					name = n
					break
				}
				c++
			}
		}
		named[name] = v.BigmapId
	}
	return named
}

// stats are stored as uint32 in a byte slice limit entrypoint count to 255
func (c *Contract) IncCallStats(entrypoint int) {
	offs := entrypoint * 4
	if cap(c.CallStats) <= offs+4 {
		// grow slice if necessary
		buf := make([]byte, offs+4)
		copy(buf, c.CallStats)
		c.CallStats = buf
	}
	c.CallStats = c.CallStats[0:util.Max(len(c.CallStats), offs+4)]
	val := binary.BigEndian.Uint32(c.CallStats[offs:])
	binary.BigEndian.PutUint32(c.CallStats[offs:], val+1)
	c.IsDirty = true
}

func (c *Contract) DecCallStats(entrypoint int) {
	offs := entrypoint * 4
	if cap(c.CallStats) <= offs+4 {
		// grow slice if necessary
		buf := make([]byte, offs+4)
		copy(buf, c.CallStats)
		c.CallStats = buf
	}
	c.CallStats = c.CallStats[0:util.Max(len(c.CallStats), offs+4)]
	val := binary.BigEndian.Uint32(c.CallStats[offs:])
	binary.BigEndian.PutUint32(c.CallStats[offs:], val-1)
	c.IsDirty = true
}

// Loads type data from already unmarshaled script or from optimized unmarshaler
func (c *Contract) LoadType() (ptyp micheline.Type, styp micheline.Type, err error) {
	if c.IsRollup() {
		err = fmt.Errorf("no script for rollup")
		return
	}
	if !c.params.IsValid() {
		if c.script != nil {
			c.params = c.script.ParamType()
			c.storage = c.script.StorageType()
		} else if c.Script != nil {
			c.params, c.storage, err = micheline.UnmarshalScriptType(c.Script)
		}
	}
	ptyp = c.params
	styp = c.storage
	return
}

// loads script and upgrades to babylon on-the-fly if originated earlier
func (c *Contract) LoadScript() (*micheline.Script, error) {
	// already cached?
	if c.script != nil {
		return c.script, nil
	}

	// pre-babylon KT1 delegtors and rollups have no script
	if len(c.Script) == 0 {
		return nil, nil
	}

	// rollups have no script
	if c.IsRollup() {
		return nil, nil
	}

	// unmarshal script
	s := micheline.NewScript()
	if err := s.UnmarshalBinary(c.Script); err != nil {
		return nil, err
	}
	c.script = s
	return s, nil
}

func (c *Contract) ConvertParams(in micheline.Parameters) (out micheline.Parameters) {
	ptyp, _, err := c.LoadType()
	if err == nil {
		ep, prim, _ := in.MapEntrypoint(ptyp)
		out = micheline.Parameters{
			Entrypoint: ep.Name,
			Value:      prim,
		}
	}
	return
}
