// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package etl

import (
	"github.com/mavryk-network/mvindex/etl/model"
	"github.com/mavryk-network/mvindex/rpc"
)

func (b *Builder) NewConstantRegistrationFlows(
	src *model.Account,
	srcbkr *model.Baker,
	fees, bal rpc.BalanceUpdates,
	id model.OpRef) []*model.Flow {

	flows, feespaid := b.NewFeeFlows(src, fees, id)

	// rest is burned
	var burned int64
	for _, u := range bal {
		if u.Kind == "contract" {
			// debit burn from source
			f := model.NewFlow(b.block, src, nil, id)
			f.Kind = model.FlowKindBalance
			f.Type = model.FlowTypeRegisterConstant
			f.AmountOut = -u.Change
			burned += -u.Change
			f.IsBurned = true
			flows = append(flows, f)
		}
	}

	// debit burn from source delegation if not baker
	if srcbkr != nil && !src.IsBaker && feespaid+burned > 0 {
		f := model.NewFlow(b.block, srcbkr.Account, src, id)
		f.Kind = model.FlowKindDelegation
		f.Type = model.FlowTypeRegisterConstant
		f.AmountOut = feespaid + burned
		flows = append(flows, f)
	}

	b.block.Flows = append(b.block.Flows, flows...)
	return flows
}

func (b *Builder) NewIncreasePaidStorageFlows(
	src *model.Account,
	srcbkr *model.Baker,
	fees, bal rpc.BalanceUpdates,
	id model.OpRef) []*model.Flow {

	flows, feespaid := b.NewFeeFlows(src, fees, id)

	// rest is burned
	var burned int64
	for _, u := range bal {
		if u.Kind == "contract" {
			// debit burn from source
			f := model.NewFlow(b.block, src, nil, id)
			f.Kind = model.FlowKindBalance
			f.Type = model.FlowTypePayStorage
			f.AmountOut = -u.Change
			burned += -u.Change
			f.IsBurned = true
			flows = append(flows, f)
		}
	}

	// debit burn from source delegation if not baker
	if srcbkr != nil && !src.IsBaker && feespaid+burned > 0 {
		f := model.NewFlow(b.block, srcbkr.Account, src, id)
		f.Kind = model.FlowKindDelegation
		f.Type = model.FlowTypePayStorage
		f.AmountOut = feespaid + burned
		flows = append(flows, f)
	}

	b.block.Flows = append(b.block.Flows, flows...)
	return flows
}
