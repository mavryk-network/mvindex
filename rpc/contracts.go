// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"

	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvgo/micheline"
)

// Contracts holds info about a Tezos account
type ContractInfo struct {
	Balance int64 `json:"balance,string"`
	// Delegate mavryk.Address `json:"delegate"`
	Counter int64 `json:"counter,string"`
}

// GetContract returns the full info about a contract at block id.
func (c *Client) GetContract(ctx context.Context, addr mavryk.Address, id BlockID) (*ContractInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", id, addr)
	var info ContractInfo
	err := c.Get(ctx, u, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetContractScript returns the originated contract script.
func (c *Client) GetContractScript(ctx context.Context, addr mavryk.Address, id BlockID) (*micheline.Script, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/script", id, addr)
	s := micheline.NewScript()
	err := c.Get(ctx, u, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetContractStorage returns the contract's storage at block id.
func (c *Client) GetContractStorage(ctx context.Context, addr mavryk.Address, id BlockID) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", id, addr)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}
