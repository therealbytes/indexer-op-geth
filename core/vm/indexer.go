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

package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type InchainIndexer struct {
	db      StateDB
	address common.Address
	slot    common.Hash
}

func NewInchainIndexer(db StateDB, slot common.Hash) *InchainIndexer {
	return &InchainIndexer{
		db:      db,
		address: params.IndexerContractAddress,
		slot:    slot,
	}
}

func (idx *InchainIndexer) getStorageKey(key common.Hash) common.Hash {
	return crypto.Keccak256Hash(key.Bytes(), idx.slot.Bytes())
}

func (idx *InchainIndexer) GetValue(key common.Hash) common.Hash {
	return idx.db.GetState(idx.address, idx.getStorageKey(key))
}

func (idx *InchainIndexer) SetValue(key common.Hash, value common.Hash) {
	idx.db.SetState(idx.address, idx.getStorageKey(key), value)
}
