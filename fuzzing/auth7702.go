package fuzzing

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

// authHelper
type authHelper struct {
	keys    map[common.Address]*ecdsa.PrivateKey
	addrs   []common.Address
	nonces  map[common.Address]uint64
	chainID *uint256.Int
}

func newHelper() *authHelper {
	h := &authHelper{
		keys:   make(map[common.Address]*ecdsa.PrivateKey),
		nonces: make(map[common.Address]uint64),
	}
	h.chainID = uint256.NewInt(0)
	addKey := func(pKey string) {
		key, err := crypto.HexToECDSA(pKey)
		if err != nil {
			panic(err)
		}
		addr := crypto.PubkeyToAddress(key.PublicKey)
		h.keys[addr] = key
		h.addrs = append(h.addrs, addr)
		log.Trace("Added key", "address", addr)
	}
	//SK   = "0x9c07c2221efd8cef3e81511f463eebe803a6feda7a27ef2dc0e25001f09248aa"
	//ADDR = "0x7a40026A3b9A41754a95EeC8c92C6B99886f440C"
	addKey("45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8")
	addKey("1111111111111111111111111111111111111111111111111111111111111111")
	addKey("2222222222222222222222222222222222222222222222222222222222222222")
	addKey("3333333333333333333333333333333333333333333333333333333333333333")
	addKey("4444444444444444444444444444444444444444444444444444444444444444")
	addKey("5555555555555555555555555555555555555555555555555555555555555555")
	addKey("6666666666666666666666666666666666666666666666666666666666666666")
	addKey("7777777777777777777777777777777777777777777777777777777777777777")
	addKey("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	addKey("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	return h
}

func (h *authHelper) consumeNonce(addr common.Address) uint64 {
	cur := h.nonces[addr]
	h.nonces[addr]++
	log.Trace("Nonce used", "address", addr, "next", h.nonces[addr])
	return cur
}

func fill7702(gst *GstMaker, fork string) {
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
	var allAddresses []common.Address
	allAddresses = append(allAddresses, contracts...)
	allAddresses = append(allAddresses, h.addrs...)
	// Add the empty-addr (acts as a clearing-marker)
	allAddresses = append(allAddresses, common.Address{})
	// Also add precompile-addresses
	allAddresses = append(allAddresses, vm.PrecompiledAddressesPrague...)

	// each contract does a bit calling within the global set
	for _, addr := range contracts {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCall2200(allAddresses),
			Balance: new(big.Int),
			Storage: RandStorage(15, 20),
		})
	}

	// make the EOAs exist too
	for _, addr := range h.addrs[1:] {
		gst.AddAccount(addr, GenesisAccount{
			Balance: big.NewInt(1),
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	//
	var authList []*stAuthorization
	for i := 0; i < 1+rand.Intn(25); i++ {
		source := h.addrs[rand.Int()%len(h.addrs)]
		dest := allAddresses[rand.Int()%len(allAddresses)]

		nonce := h.consumeNonce(source)
		unsigned := types.SetCodeAuthorization{
			ChainID: *h.chainID,
			Address: dest,
			Nonce:   nonce,
		}
		switch rand.Intn(20) {
		case 0:
			// Random chain id
			unsigned.ChainID = randU256()
		case 1:
			// Random nonce
			unsigned.Nonce = rand.Uint64()
		}
		a, err := types.SignSetCode(h.keys[source], unsigned)
		//		a, err := h.makeAuth(source, dest)
		if err != nil {
			panic(err)
		}
		authList = append(authList, &stAuthorization{
			ChainID: a.ChainID.ToBig(),
			Address: a.Address,
			Nonce:   a.Nonce,
			V:       a.V,
			R:       a.R.ToBig(),
			S:       a.S.ToBig(),
			Signer:  &source,
		})
	}

	// The transaction
	{
		tx := &StTransaction{
			// 8M gaslimit
			GasLimit:             []uint64{8000000},
			Nonce:                0,
			Value:                []string{randHex(4)},
			Data:                 []string{randHex(100)},
			MaxFeePerGas:         big.NewInt(0x10),
			MaxPriorityFeePerGas: big.NewInt(0x10),
			To:                   allAddresses[rand.Int()%len(allAddresses)].Hex(),
			Sender:               sender,
			PrivateKey:           hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
			AuthorizationList:    authList,
		}
		gst.SetTx(tx)
	}
}

func randU256() uint256.Int {
	var a uint256.Int
	if rand.Int()%2 == 0 {
		a[0] = rand.Uint64()
	}
	if rand.Int()%2 == 0 {
		a[1] = rand.Uint64()
	}
	if rand.Int()%2 == 0 {
		a[2] = rand.Uint64()
	}
	if rand.Int()%2 == 0 {
		a[3] = rand.Uint64()
	}
	return a
}
