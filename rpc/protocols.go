// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import "github.com/mavryk-network/mvgo/mavryk"

var (
	ParseProto = mavryk.MustParseProtocolHash
	ParseChain = mavryk.MustParseChainIdHash
)

var (
	ProtoAlpha     = ParseProto("ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK")
	ProtoGenesis   = ParseProto("PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i")
	ProtoBootstrap = ParseProto("Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P")
	ProtoV001      = ParseProto("PtAtLasomUEW99aVhVTrqjCHjJSpFUa8uHNEAEamx9v2SNeTaNp")

	// aliases
	PtAtLas = ProtoV001

	Mainnet  = ParseChain("NetXdQprcVkpaWU")
	Ghostnet = ParseChain("NetXnHfVqm9iesp")
	Atlasnet = ParseChain("NetXvyTAafh8goH")

	Versions = map[mavryk.ProtocolHash]int{
		ProtoGenesis:   0,
		ProtoBootstrap: 0,
		ProtoV001:      18,
		ProtoAlpha:     19,
	}
)
