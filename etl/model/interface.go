// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package model

import (
	"context"
	"time"

	"blockwatch.cc/packdb/pack"
	"github.com/mavryk-network/tzgo/micheline"
	"github.com/mavryk-network/tzgo/tezos"
	"github.com/mavryk-network/tzindex/etl/task"
	"github.com/mavryk-network/tzindex/rpc"
)

// Model is the interface all data models must implement.
type Model interface {
	TableKey() string
	TableOpts() pack.Options
	IndexOpts(string) pack.Options
}

// BlockCrawler is an interface to access the block crawler state and execute DB queries.
type BlockCrawler interface {
	// returns the requested database table if exists or error otherwise
	Table(string) (*pack.Table, error)

	// returns the blockchain params at specified block height
	ParamsByHeight(height int64) *rpc.Params

	// returns the blockchain params for the specified protocol
	ParamsByProtocol(proto tezos.ProtocolHash) *rpc.Params

	// returns the current crawler chain tip
	Tip() *ChainTip

	// returns the crawler's most recently seen block height
	Height() int64

	// returns stored (main chain) block at specified height
	BlockByHeight(ctx context.Context, height int64) (*Block, error)

	// returns stored chain data at specified height
	ChainByHeight(ctx context.Context, height int64) (*Chain, error)

	// returns stored supply table data at specified height
	SupplyByHeight(ctx context.Context, height int64) (*Supply, error)

	// returns height for timestamp
	BlockHeightFromTime(ctx context.Context, tm time.Time) int64
}

// BlockBuilder is an interface to access block builder caches and info.
type BlockBuilder interface {
	// resolves account from builder cache, returns nil and false when not found
	AccountByAddress(tezos.Address) (*Account, bool)

	// resolves account from database, returns nil and error when not found
	LoadAccountByAddress(context.Context, tezos.Address) (*Account, error)

	// resolves account from id, returns nil and false when not found
	AccountById(AccountID) (*Account, bool)

	// resolves baker from address, returns nil and false when not found
	BakerByAddress(tezos.Address) (*Baker, bool)

	// resolves baker from id, returns nil and false when not found
	BakerById(AccountID) (*Baker, bool)

	// resolves contract from account id, returns nil and false when not found
	ContractById(AccountID) (*Contract, bool)

	// returns a map of all accounts referenced in the current block
	Accounts() map[AccountID]*Account

	// returns a map of all delegates referenced in the current block
	Bakers() map[AccountID]*Baker

	// returns a map of all contracts referenced in the current block
	Contracts() map[AccountID]*Contract

	// returns a map of all constants referenced in the current block
	Constants() micheline.ConstantDict

	// return params at specific height
	Params(int64) *rpc.Params

	// returns the requested database table if exists or error otherwise
	Table(string) (*pack.Table, error)

	// returns global task scheduler
	Sched() *task.Scheduler

	// returns true if indexer is run in light mode
	IsLightMode() bool
}

// BlockIndexer is the interface all indexers must implement.
type BlockIndexer interface {
	// Name returns the human-readable name of the index.
	Name() string

	// Key returns the key of the index as a string.
	Key() string

	// Create is invoked when the indexer manager determines the index needs
	// to be created for the first time.
	Create(path, label string, opts interface{}) error

	// Init is invoked when the table manager is first initializing the
	// datastore.  This differs from the Create method in that it is called on
	// every load, including the case the datatable was just created.
	Init(path, label string, opts interface{}) error

	// ConnectBlock is invoked when the table manager is notified that a new
	// block has been connected to the main chain.
	ConnectBlock(ctx context.Context, block *Block, builder BlockBuilder) error

	// DisconnectBlock is invoked when the table manager is notified that a
	// block has been disconnected from the main chain.
	DisconnectBlock(ctx context.Context, block *Block, builder BlockBuilder) error

	// DeleteBlock is invoked when the table manager is notified that a
	// block must be rolled back after an error occured.
	DeleteBlock(ctx context.Context, height int64) error

	// DeleteCycle is invoked when an index must delete all content from
	// a particular cycle.
	DeleteCycle(ctx context.Context, cycle int64) error

	// FinalizeSync is invoked when an the initial sync has finished. It may
	// be used to clean and defrag tables or (re)build indexes.
	FinalizeSync(ctx context.Context) error

	// OnTaskComplete is called when a scheduled task completes.
	OnTaskComplete(context.Context, *task.TaskResult) error

	// Flush flushes all indexer databases.
	Flush(ctx context.Context) error

	// Close closes the indexer and frees all associated resources, if any.
	Close() error

	// returns the database storing all indexer tables
	DB() *pack.DB

	// returns the list of database tables used by the indexer
	Tables() []*pack.Table
}
