// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package etl

import (
	"blockwatch.cc/packdb/pack"
	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvindex/etl/model"
)

type ListRequest struct {
	Account     *model.Account
	Mode        pack.FilterMode
	Typs        model.OpTypeList
	Since       int64
	Until       int64
	Offset      uint
	Limit       uint
	Cursor      uint64
	Order       pack.OrderType
	SenderId    model.AccountID
	ReceiverId  model.AccountID
	Entrypoints []int64
	Period      int64
	BigmapId    int64
	BigmapKey   mavryk.ExprHash
	OpId        model.OpID
	WithStorage bool
}

func (r ListRequest) WithDelegation() bool {
	if r.Mode == pack.FilterModeEqual || r.Mode == pack.FilterModeIn {
		for _, t := range r.Typs {
			if t == model.OpTypeDelegation {
				return true
			}
		}
		return false
	} else {
		for _, t := range r.Typs {
			if t == model.OpTypeDelegation {
				return false
			}
		}
		return true
	}
}
