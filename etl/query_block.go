// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package etl

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"blockwatch.cc/packdb/pack"
	"github.com/mavryk-network/mvgo/mavryk"
	"github.com/mavryk-network/mvindex/etl/model"
)

func (m *Indexer) BlockByID(ctx context.Context, id uint64) (*model.Block, error) {
	if id == 0 {
		return nil, model.ErrNoBlock
	}
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return nil, err
	}
	b := &model.Block{}
	err = pack.NewQuery("api.block_by_parent_id").
		WithTable(table).
		AndEqual("I", id).
		Execute(ctx, b)
	if err != nil {
		return nil, err
	}
	if b.RowId == 0 {
		return nil, model.ErrNoBlock
	}
	b.Params, _ = m.reg.GetParamsByDeployment(b.Version)
	return b, nil
}

// find a block's canonical successor (non-orphan)
func (m *Indexer) BlockByParentId(ctx context.Context, id uint64) (*model.Block, error) {
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return nil, err
	}
	b := &model.Block{}
	err = pack.NewQuery("api.block_by_parent_id").
		WithTable(table).
		AndEqual("parent_id", id).
		WithLimit(1).
		Execute(ctx, b)
	if err != nil {
		return nil, err
	}
	if b.RowId == 0 {
		return nil, model.ErrNoBlock
	}
	b.Params, _ = m.reg.GetParamsByDeployment(b.Version)
	return b, nil
}

func (m *Indexer) BlockHashByHeight(ctx context.Context, height int64) (mavryk.BlockHash, error) {
	type XBlock struct {
		Hash mavryk.BlockHash `pack:"H"`
	}
	b := &XBlock{}
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return b.Hash, err
	}
	err = pack.NewQuery("api.block_hash_by_height").
		WithTable(table).
		AndEqual("height", height).
		WithLimit(1).
		Execute(ctx, b)
	if err != nil {
		return b.Hash, err
	}
	if !b.Hash.IsValid() {
		return b.Hash, model.ErrNoBlock
	}
	return b.Hash, nil
}

func (m *Indexer) BlockHashById(ctx context.Context, id uint64) (mavryk.BlockHash, error) {
	type XBlock struct {
		Hash mavryk.BlockHash `pack:"H"`
	}
	b := &XBlock{}
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return b.Hash, err
	}
	err = pack.NewQuery("api.block_hash_by_id").
		WithTable(table).
		WithFields("H").
		AndEqual("I", id).
		Execute(ctx, b)
	if err != nil {
		return b.Hash, err
	}
	if !b.Hash.IsValid() {
		return b.Hash, model.ErrNoBlock
	}
	return b.Hash, nil
}

func (m *Indexer) BlockByHeight(ctx context.Context, height int64) (*model.Block, error) {
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return nil, err
	}
	b := &model.Block{}
	err = pack.NewQuery("api.block_by_height").
		WithTable(table).
		AndEqual("height", height).
		Execute(ctx, b)
	if err != nil {
		return nil, err
	}
	if b.RowId == 0 {
		return nil, model.ErrNoBlock
	}
	b.Params, _ = m.reg.GetParamsByDeployment(b.Version)
	return b, nil
}

func (m *Indexer) BlockByHash(ctx context.Context, h mavryk.BlockHash, from, to int64) (*model.Block, error) {
	if !h.IsValid() {
		return nil, fmt.Errorf("invalid block hash %s", h)
	}
	table, err := m.Table(model.BlockTableKey)
	if err != nil {
		return nil, err
	}
	q := pack.NewQuery("api.block_by_hash").
		WithTable(table).
		WithLimit(1).
		WithDesc()
	if from > 0 {
		q = q.AndGte("height", from)
	}
	if to > 0 {
		q = q.AndLte("height", to)
	}
	// most expensive condition last
	q = q.AndEqual("hash", h[:])
	b := &model.Block{}
	if err = q.Execute(ctx, b); err != nil {
		return nil, err
	}
	if b.RowId == 0 {
		return nil, model.ErrNoBlock
	}
	b.Params, _ = m.reg.GetParamsByDeployment(b.Version)
	return b, nil
}

func (m *Indexer) LookupBlockId(ctx context.Context, blockIdent string) (mavryk.BlockHash, int64, error) {
	var err error
	switch {
	case blockIdent == "head":
		if b, err2 := m.BlockByHeight(ctx, m.tips[model.BlockTableKey].Height); err2 == nil {
			return b.Hash, b.Height, nil
		} else {
			err = err2
		}
	case len(blockIdent) == mavryk.HashTypeBlock.B58Len || strings.HasPrefix(blockIdent, mavryk.HashTypeBlock.B58Prefix):
		// assume it's a hash
		if blockHash, err2 := mavryk.ParseBlockHash(blockIdent); err2 == nil && blockHash.IsValid() {
			if b, err3 := m.BlockByHash(ctx, blockHash, 0, 0); err3 == nil {
				return b.Hash, b.Height, nil
			} else {
				err = err3
			}
		} else {
			err = model.ErrInvalidBlockHash
		}
	default:
		// try parsing as height
		if h, err2 := strconv.ParseUint(blockIdent, 10, 64); err2 == nil {
			return m.LookupBlockHash(ctx, int64(h)), int64(h), nil
		}
		err = model.ErrInvalidBlockHeight
	}
	return mavryk.BlockHash{}, 0, err
}

func (m *Indexer) LookupBlock(ctx context.Context, blockIdent string) (*model.Block, error) {
	var (
		b   *model.Block
		err error
	)
	switch {
	case blockIdent == "head":
		b, err = m.BlockByHeight(ctx, m.tips[model.BlockTableKey].Height)
	case len(blockIdent) == mavryk.HashTypeBlock.B58Len || strings.HasPrefix(blockIdent, mavryk.HashTypeBlock.B58Prefix):
		// assume it's a hash
		var blockHash mavryk.BlockHash
		blockHash, err = mavryk.ParseBlockHash(blockIdent)
		if err != nil || !blockHash.IsValid() {
			return nil, model.ErrInvalidBlockHash
		}
		b, err = m.BlockByHash(ctx, blockHash, 0, 0)
	default:
		// try parsing as height
		if h, err2 := strconv.ParseUint(blockIdent, 10, 64); err2 != nil {
			return nil, model.ErrInvalidBlockHeight
		} else {
			b, err = m.BlockByHeight(ctx, int64(h))
		}
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (m *Indexer) LookupLastBakedBlock(ctx context.Context, bkr *model.Baker, height int64) (int64, error) {
	if bkr.BlocksBaked == 0 {
		return 0, model.ErrNoBlock
	}
	t, err := m.Table(model.RightsTableKey)
	if err != nil {
		return 0, err
	}
	right := model.Right{}
	err = pack.NewQuery("last_bake").
		WithTable(t).
		WithDesc().
		WithFields("height", "blocks_baked").
		AndEqual("account_id", bkr.AccountId).
		AndLte("height", height).
		Stream(ctx, func(r pack.Row) error {
			if err := r.Decode(&right); err != nil {
				return err
			}
			if right.Baked.Count() == 0 {
				height = right.Height - 1
				return nil
			}
			pos := int(height - right.Height)
			next, _ := right.Baked.Reverse().Run(pos)
			if next < 0 {
				height = right.Height - 1
				return nil
			} else {
				height = right.Height + int64(next)
				return io.EOF
			}
		})
	if err != nil && err != io.EOF {
		return 0, err
	}
	return height, nil
}

func (m *Indexer) LookupLastEndorsedBlock(ctx context.Context, bkr *model.Baker, height int64) (int64, error) {
	if bkr.BlocksEndorsed == 0 {
		return 0, model.ErrNoBlock
	}
	t, err := m.Table(model.RightsTableKey)
	if err != nil {
		return 0, err
	}
	right := model.Right{}
	err = pack.NewQuery("last_endorse").
		WithTable(t).
		WithDesc().
		WithFields("height", "blocks_endorsed").
		AndEqual("account_id", bkr.AccountId).
		AndLte("height", height).
		Stream(ctx, func(r pack.Row) error {
			if err := r.Decode(&right); err != nil {
				return err
			}
			if right.Endorsed.Count() == 0 {
				height = right.Height - 1
				return nil
			}
			pos := int(height - right.Height)
			next, _ := right.Endorsed.Reverse().Run(pos)
			if next < 0 {
				height = right.Height - 1
				return nil
			} else {
				height = right.Height + int64(next)
				return io.EOF
			}
		})
	if err != nil && err != io.EOF {
		return 0, err
	}
	return height, nil
}

func (m *Indexer) ListBlockRights(ctx context.Context, height int64, typ mavryk.RightType) ([]model.BaseRight, error) {
	p := m.ParamsByHeight(height)
	pos := int(p.CyclePosition(height))

	table, err := m.Table(model.RightsTableKey)
	if err != nil {
		return nil, err
	}
	q := pack.NewQuery("api.list_rights").
		WithTable(table).
		AndEqual("cycle", p.HeightToCycle(height))
	if typ.IsValid() {
		q = q.AndEqual("type", typ)
	}
	resp := make([]model.BaseRight, 0)
	right := model.Right{}
	err = q.Stream(ctx, func(r pack.Row) error {
		if err := r.Decode(&right); err != nil {
			return err
		}
		switch typ {
		case mavryk.RightTypeBaking:
			if r, ok := right.ToBase(pos, mavryk.RightTypeBaking); ok {
				resp = append(resp, r)
			}
		case mavryk.RightTypeEndorsing:
			if r, ok := right.ToBase(pos, mavryk.RightTypeEndorsing); ok {
				resp = append(resp, r)
			}
		default:
			if r, ok := right.ToBase(pos, mavryk.RightTypeBaking); ok {
				resp = append(resp, r)
			}
			if r, ok := right.ToBase(pos, mavryk.RightTypeEndorsing); ok {
				resp = append(resp, r)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Indexer) CycleByNum(ctx context.Context, num int64) (*model.Cycle, error) {
	table, err := m.Table(model.CycleTableKey)
	if err != nil {
		return nil, err
	}
	c := &model.Cycle{}
	err = pack.NewQuery("api.cycle_by_num").
		WithTable(table).
		AndEqual("cycle", num).
		Execute(ctx, c)
	if err != nil {
		return nil, err
	}
	if c.RowId == 0 {
		return nil, model.ErrNoCycle
	}
	return c, nil
}
