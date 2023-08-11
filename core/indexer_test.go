// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

//go:embed callerBytecode.txt
var callerBytecodeHex string

//go:embed caller.abi
var callerABI string

func TestIndexer(t *testing.T) {
	var (
		key, _         = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		senderAddress  = crypto.PubkeyToAddress(key.PublicKey)
		contract0      = common.HexToAddress("0xc0ffee0000000000000000000000000000000000")
		contract1      = common.HexToAddress("0xc0ffee0000000000000000000000000000000001")
		contract2      = common.HexToAddress("0xc0ffee0000000000000000000000000000000002")
		contract3      = common.HexToAddress("0xc0ffee0000000000000000000000000000000003")
		callerBytecode = common.Hex2Bytes(callerBytecodeHex)
		gspec          = &Genesis{
			GasLimit: 2e7,
			Config:   params.TestChainConfig,
			Alloc: GenesisAlloc{
				senderAddress: {Balance: big.NewInt(1e18)},
				contract0:     {Balance: common.Big0, Code: callerBytecode},
				contract1:     {Balance: common.Big0, Code: callerBytecode},
				contract2:     {Balance: common.Big0, Code: callerBytecode},
				contract3:     {Balance: common.Big0, Code: callerBytecode},
			},
		}
		signer     = types.LatestSigner(gspec.Config)
		txGasLimit = uint64(1e7)
	)

	abiReader := strings.NewReader(callerABI)
	ABI, err := abi.JSON(abiReader)
	if err != nil {
		panic(err)
	}

	db, blocks, receipts := GenerateChainWithGenesis(gspec, ethash.NewFaker(), 1, func(ii int, block *BlockGen) {
		data, err := ABI.Pack("call", [][]common.Address{{contract1, contract2}, {contract3}}, common.Big0)
		if err != nil {
			panic(err)
		}
		tx := types.NewTransaction(block.TxNonce(senderAddress), contract0, common.Big0, txGasLimit, block.BaseFee(), data)
		signed, err := types.SignTx(tx, signer, key)
		if err != nil {
			panic(err)
		}
		block.AddTx(signed)
	})

	for _, rr := range receipts {
		for _, receipt := range rr {
			if receipt.Status != types.ReceiptStatusSuccessful {
				data, err := json.MarshalIndent(receipt, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(data))
				panic("receipt status not successful")
			}
			for _, log := range receipt.Logs {
				fmt.Println(log.Address)
			}
		}
	}

	root := blocks[len(blocks)-1].Root()
	statedb, err := state.New(root, state.NewDatabase(db), nil)
	if err != nil {
		panic(err)
	}
	gasIndexer := vm.NewInchainIndexer(statedb, common.BigToHash(common.Big0))
	fmt.Println("gas0:", gasIndexer.GetValue(contract0.Hash()).Big().Uint64())
	fmt.Println("gas1:", gasIndexer.GetValue(contract1.Hash()).Big().Uint64())
	fmt.Println("gas2:", gasIndexer.GetValue(contract2.Hash()).Big().Uint64())
	fmt.Println("gas3:", gasIndexer.GetValue(contract3.Hash()).Big().Uint64())
}
