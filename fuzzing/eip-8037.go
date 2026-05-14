// Copyright Martin Holst Swende
// Copyright 2026 Spencer Taylor-Brown (terminus-31)
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package fuzzing

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

// gasLimitOptions8037 enumerates tx.gas_limit values that exercise the
// EIP-8037 state-gas reservoir at and around the EIP-7825 cap
// (TX_MAX_GAS_LIMIT = 2**24 = 16,777,216). For tx.gas_limit > cap, the
// excess seeds the reservoir; state-gas charges then draw from
// reservoir first and spill into gas_left when empty.
var gasLimitOptions8037 = []uint64{
	8_000_000,   // well below cap, regular gas only, no reservoir
	16_000_000,  // near cap, no reservoir yet
	16_777_215,  // 1 below cap, last possible no-reservoir tx
	16_777_216,  // at cap exactly, reservoir = 0
	16_777_217,  // 1 above cap, reservoir = 1 (minimum spillover)
	20_000_000,  // ~3.2M reservoir, moderate state-gas budget
	30_000_000,  // ~13.2M reservoir, exercises spillover paths
	50_000_000,  // ~33.2M reservoir, large state-gas surplus
}

// rand8037AuthList builds an EIP-7702 authorization list of 1-5 entries.
// Target is biased across four refund pathways the EIP defines:
//   - clearing (Address=0x0): existing-account refund + delegation-clear refund
//   - precompile (0x01): existing-account refund (precompiles always exist)
//   - self (source==dest): existing-account refund when source pre-existed
//   - random (other): refund only if target pre-existed
// Validity is occasionally corrupted (~5% wrong chainID, ~5% wrong nonce)
// to hit the "invalid auth still charges intrinsic" path.
func rand8037AuthList(h *authHelper, allAddrs []common.Address) []*stAuthorization {
	var list []*stAuthorization
	for i := 0; i < 1+rand.Intn(5); i++ {
		source := h.addrs[rand.Intn(len(h.addrs))]
		var dest common.Address
		switch rand.Intn(4) {
		case 0:
			dest = common.Address{} // clearing delegation
		case 1:
			dest = common.HexToAddress("0x01") // precompile (ecrecover)
		case 2:
			dest = source // self
		default:
			dest = allAddrs[rand.Intn(len(allAddrs))]
		}
		nonce := h.consumeNonce(source)
		unsigned := types.SetCodeAuthorization{
			ChainID: *h.chainID,
			Address: dest,
			Nonce:   nonce,
		}
		switch rand.Intn(20) {
		case 0:
			unsigned.ChainID = randU256() // ~5% wrong chainID
		case 1:
			unsigned.Nonce = rand.Uint64() // ~5% wrong nonce
		}
		a, err := types.SignSetCode(h.keys[source], unsigned)
		if err != nil {
			panic(err)
		}
		list = append(list, &stAuthorization{
			ChainID: a.ChainID.ToBig(),
			Address: a.Address,
			Nonce:   a.Nonce,
			V:       a.V,
			R:       a.R.ToBig(),
			S:       a.S.ToBig(),
			Signer:  &source,
		})
	}
	return list
}

// fill8037 seeds a state test that exercises EIP-8037 (State Creation
// Gas Cost Increase) surface area under the Amsterdam fork rules.
//
// Generator strategy:
//   - Bytecode via RandCall2200: ~10% SSTORE (random values in [0..3]
//     across random slots produce 0→n / n→0 / n→m / n→n transitions),
//     ~10% CREATE/CREATE2, ~5% SELFDESTRUCT, plus random nested calls
//     (recursive up to depth 10), returns and reverts. Under Amsterdam
//     every state-touching op draws on the state-gas reservoir.
//   - Gas limit randomised across the reservoir boundary (see
//     gasLimitOptions8037) — cardinal value is 16,777,216
//     (TX_MAX_GAS_LIMIT); below it reservoir is zero, above it the
//     excess seeds the reservoir and is consumed first.
//   - EIP-7702 authorization list attached to ~50% of txs (1-5
//     entries), with bias across target types (clearing / precompile /
//     self / random) and occasional invalid nonce/chainID — exercises
//     every refund pathway in the EIP-7702 × EIP-8037 surface.
//   - Helper EOA pre-state randomised per-account (50% pre-existing
//     with balance=1, 50% nonexistent) so auth signers hit both the
//     "existing-account refund" and the "new-account no-refund" paths.
//
// Out of scope here (better as targeted sub-engines, see
// eip-8037-*.go):
//   - Deterministic SSTORE 0→x→0 restoration sequences
//   - Engineered deep call chains (depth 20-50)
//   - Same-tx CREATE+SELFDESTRUCT no-refund path
func fill8037(gst *GstMaker, fork string) {
	h := newHelper()
	contracts := []common.Address{
		common.HexToAddress("0xF1"),
		common.HexToAddress("0xF2"),
		common.HexToAddress("0xF3"),
		common.HexToAddress("0xF4"),
		common.HexToAddress("0xF5"),
		common.HexToAddress("0xF6"),
		common.HexToAddress("0xF7"),
		common.HexToAddress("0xF8"),
		common.HexToAddress("0xF9"),
		common.HexToAddress("0xFA"),
	}
	nonGenesisAddresses := []common.Address{
		common.HexToAddress("0x00"),
		common.HexToAddress("0x01"),
		common.HexToAddress("0x02"),
		common.HexToAddress("0x03"),
		common.HexToAddress("0x04"),
		common.HexToAddress("0x05"),
		common.HexToAddress("0x06"),
		common.HexToAddress("0x07"),
		common.HexToAddress("0x08"),
		common.HexToAddress("0x09"),
		common.HexToAddress("0x0A"),
		common.HexToAddress("0x0B"),
		common.HexToAddress("0x0C"),
		common.HexToAddress("0x0D"),
		common.HexToAddress("0x0E"),
	}
	var allAddrs []common.Address
	allAddrs = append(allAddrs, contracts...)
	allAddrs = append(allAddrs, nonGenesisAddresses...)
	allAddrs = append(allAddrs, h.addrs...)

	// Pre-state: small balances at nonGenesis addresses, contracts with
	// random storage and RandCall2200 bytecode.
	for _, addr := range nonGenesisAddresses {
		gst.AddAccount(addr, GenesisAccount{
			Balance: new(big.Int).SetUint64(1),
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	for _, addr := range contracts {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCall2200(allAddrs),
			Balance: new(big.Int),
			Storage: RandStorage(15, 20),
		})
	}
	// Helper EOAs: 50/50 mix of pre-existing (balance=1) vs nonexistent
	// (not added to genesis). Auths from a pre-existing source hit the
	// 183,600 existing-account refund path; auths from a nonexistent
	// source charge the full 218,790 intrinsic with no refund.
	for _, addr := range h.addrs[1:] {
		if rand.Intn(2) == 0 {
			gst.AddAccount(addr, GenesisAccount{
				Balance: big.NewInt(1),
				Storage: make(map[common.Hash]common.Hash),
			})
		}
	}

	// Authorization list: ~50% chance, 1-5 entries with varied targets
	// and occasional invalid nonce/chainID (see rand8037AuthList).
	var authList []*stAuthorization
	if rand.Intn(2) == 0 {
		authList = rand8037AuthList(h, allAddrs)
	}

	gasLimit := gasLimitOptions8037[rand.Intn(len(gasLimitOptions8037))]
	tx := &StTransaction{
		GasLimit:             []uint64{gasLimit},
		Nonce:                0,
		Value:                []string{randHex(4)},
		Data:                 []string{randHex(100)},
		MaxFeePerGas:         big.NewInt(0x10),
		MaxPriorityFeePerGas: big.NewInt(0x10),
		To:                   contracts[0].Hex(),
		Sender:               sender,
		PrivateKey:           hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		AuthorizationList:    authList,
	}
	gst.SetTx(tx)
}
