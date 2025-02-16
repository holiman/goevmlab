package fuzzing

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

// authHelper
type authHelper struct {
	keys    map[common.Address]*ecdsa.PrivateKey
	addrs   []common.Address
	nonces  map[common.Address]uint64
	chainId *uint256.Int
}

func newHelper() *authHelper {
	h := &authHelper{
		keys:   make(map[common.Address]*ecdsa.PrivateKey),
		nonces: make(map[common.Address]uint64),
	}
	h.chainId = uint256.NewInt(1)
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

// makes an auth using the current nonce, and bumps the nonce
func (h *authHelper) makeAuth(from, to common.Address) (types.SetCodeAuthorization, error) {
	nonce := h.consumeNonce(from)
	return h.makeAuthWithNonce(from, nonce, to)
}

func (h *authHelper) makeAuthWithNonce(from common.Address, nonce uint64, to common.Address) (types.SetCodeAuthorization, error) {
	unsigned := types.SetCodeAuthorization{
		ChainID: *h.chainId,
		Address: to,
		Nonce:   nonce,
	}
	return types.SignSetCode(h.keys[from], unsigned)
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
		dest := h.addrs[rand.Int()%len(h.addrs)]
		a, err := h.makeAuth(source, dest)
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
		})
	}

	// The transaction
	{
		tx := &StTransaction{
			// 8M gaslimit
			GasLimit:          []uint64{8000000},
			Nonce:             0,
			Value:             []string{randHex(4)},
			Data:              []string{randHex(100)},
			GasPrice:          big.NewInt(0x10),
			To:                allAddresses[rand.Int()%len(allAddresses)].Hex(),
			Sender:            sender,
			PrivateKey:        hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
			AuthorizationList: authList,
		}
		gst.SetTx(tx)
	}
}
