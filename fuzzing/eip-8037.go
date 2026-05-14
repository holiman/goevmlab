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

// fill8037 seeds a state test that exercises EIP-8037 (State Creation
// Gas Cost Increase) surface area under the Amsterdam fork rules.
//
// Generator strategy:
//   - Bytecode via RandCall2200: ~10% SSTORE, ~10% CREATE/CREATE2,
//     ~5% SELFDESTRUCT, plus random calls, returns, and reverts —
//     every state-touching op now draws on the state-gas reservoir
//     under Amsterdam, so this covers the broad surface.
//   - Gas limit randomised across the reservoir boundary (see
//     gasLimitOptions8037) — the cardinal value is 16,777,216
//     (TX_MAX_GAS_LIMIT); below it the reservoir is zero, above it
//     the excess seeds the reservoir and is consumed first.
//
// Follow-ups can target EIP-7702 auth-state refunds, revert-vs-commit
// splits, and reservoir-exact-equal-to-needed boundary cases.
func fill8037(gst *GstMaker, fork string) {
	addrs := []common.Address{
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
	allAddrs = append(allAddrs, addrs...)
	allAddrs = append(allAddrs, nonGenesisAddresses...)
	for _, addr := range nonGenesisAddresses {
		gst.AddAccount(addr, GenesisAccount{
			Balance: new(big.Int).SetUint64(1),
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCall2200(allAddrs),
			Balance: new(big.Int),
			Storage: RandStorage(15, 20),
		})
	}
	{
		gasLimit := gasLimitOptions8037[rand.Intn(len(gasLimitOptions8037))]
		tx := &StTransaction{
			GasLimit:   []uint64{gasLimit},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x10),
			To:         addrs[0].Hex(),
			Sender:     sender,
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		gst.SetTx(tx)
	}
}
