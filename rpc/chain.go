// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"

	"github.com/mavryk-network/mvgo/mavryk"
)

// GetChainId returns the chain id (i.e. network id).
// https://mavryk.gitlab.io/shell/rpc.html#get-chains-chain-id-chain-id
func (c *Client) GetChainId(ctx context.Context) (mavryk.ChainIdHash, error) {
	var id mavryk.ChainIdHash
	err := c.Get(ctx, "chains/main/chain_id", &id)
	return id, err
}

type Status struct {
	Bootstrapped bool   `json:"bootstrapped"`
	SyncState    string `json:"sync_state"`
}

// GetStatus returns whether the node is bootstrapped (i.e. has downloaded
// the full chain) and in sync.
// https://mavryk.gitlab.io/shell/rpc.html#get-chains-chain-id-is-bootstrapped
func (c *Client) GetStatus(ctx context.Context) (Status, error) {
	var s Status
	err := c.Get(ctx, "chains/main/is_bootstrapped", &s)
	return s, err
}

type NodeVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	// AdditionalInfo string `json:"additional_info"` // v015
	// AdditionalInfo map[string]any `json:"additional_info"` // v016+
}

type NetworkVersion struct {
	ChainName            string `json:"chain_name"`
	DistributedDbVersion int    `json:"distributed_db_version"`
	P2pVersion           int    `json:"p2p_version"`
}

type CommitInfo struct {
	CommitHash string `json:"commit_hash"`
	CommitDate string `json:"commit_date"`
}

type VersionInfo struct {
	NodeVersion    NodeVersion    `json:"version"`
	NetworkVersion NetworkVersion `json:"network_version"`
	CommitInfo     CommitInfo     `json:"commit_info"`
}

// GetVersion returns node's version info.
// https://mavryk.gitlab.io/shell/rpc.html#get-version
func (c *Client) GetVersionInfo(ctx context.Context) (VersionInfo, error) {
	var v VersionInfo
	err := c.Get(ctx, "version", &v)
	return v, err
}
