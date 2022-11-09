package blockchain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/nbright/nomadcoin/db"
	"github.com/nbright/nomadcoin/utils"
)

// func (b *BlockChain) listnewBlocks() {
// 	for _, block := range b.newBlocks {
// 		fmt.Printf("Data: %s\n", block.data)
// 		fmt.Printf("Hash: %s\n", block.hash)
// 		fmt.Printf("PreHash: %s\n", block.prevHash)
// 	}
// }

const (
	defaultDifficulty  int = 2
	difficultyInterval int = 5
	blockInterval      int = 2
	allowedRange       int = 2
)

type blockChain struct {
	NewestHash        string `json:"newestHash"`
	Height            int    `json:"height"`
	CurrentDifficulty int    `json:"currentDifficulty"`
	m                 sync.Mutex
}

var b *blockChain
var once sync.Once

func (b *blockChain) restore(data []byte) {
	utils.FromBytes(b, data)
}
func (b *blockChain) AddBlock() *Block {
	block := createBlock(b.NewestHash, b.Height+1, getDifficulty(b))
	b.NewestHash = block.Hash
	b.Height = block.Height
	b.CurrentDifficulty = block.Difficulty
	persistBlockchain(b)
	return block
}

func persistBlockchain(b *blockChain) {
	db.SaveBlockchain(utils.ToBytes(b))
}

func Blocks(b *blockChain) []*Block {
	b.m.Lock()
	defer b.m.Unlock()
	var newBlocks []*Block
	hashCursor := b.NewestHash
	for {
		block, _ := FindBlock(hashCursor)
		newBlocks = append(newBlocks, block)

		if block.PrevHash != "" {
			hashCursor = block.PrevHash
		} else {
			break
		}
	}
	return newBlocks
}

func Txs(b *blockChain) []*Tx {
	var txs []*Tx
	for _, block := range Blocks(b) {
		txs = append(txs, block.Transactions...)
	}
	return txs
}

func FindTx(b *blockChain, targetID string) *Tx {
	for _, tx := range Txs(b) {
		if tx.Id == targetID {
			return tx
		}
	}
	return nil

}
func recalculateDifficulty(b *blockChain) int {
	allBlock := Blocks(b)
	newestBlock := allBlock[0]
	lastRecalculatedBlock := allBlock[difficultyInterval-1]
	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculatedBlock.Timestamp / 60)
	expectedTime := difficultyInterval * blockInterval
	if actualTime < (expectedTime - allowedRange) {
		return b.CurrentDifficulty + 1
	} else if actualTime > (expectedTime + allowedRange) {
		return b.CurrentDifficulty - 1
	}
	return b.CurrentDifficulty
}

func getDifficulty(b *blockChain) int {
	if b.Height == 0 {
		return defaultDifficulty
	} else if b.Height%difficultyInterval == 0 {
		// recalculate the difficulty
		return recalculateDifficulty(b)
	} else {
		return b.CurrentDifficulty
	}
}

/* 변경
func (b *blockChain) txOuts() []*TxOut {
	var txOuts []*TxOut
	newBlocks := b.newBlocks()
	for _, block := range newBlocks {
		for _, tx := range block.Transactions {
			txOuts = append(txOuts, tx.TxOuts...)
		}
	}
	return txOuts
}
*/
// UnSpent transaction outputs
func UTxOutsByAddress(address string, b *blockChain) []*UTxOut {
	/*
		var ownedTxOuts []*TxOut
		txOuts := b.txOuts()
		for _, txOut := range txOuts {
			if txOut.Owner == address {
				ownedTxOuts = append(ownedTxOuts, txOut)
			}
		}
		return ownedTxOuts
	*/
	var uTxOuts []*UTxOut
	creatorTxs := make(map[string]bool)
	for _, block := range Blocks(b) {
		for _, tx := range block.Transactions {
			for _, input := range tx.TxIns {
				if input.Signature == "COINBASE" {
					break
				}
				if FindTx(b, input.TxID).TxOuts[input.Index].Address == input.Signature {
					creatorTxs[input.TxID] = true
				}
			}
			for index, output := range tx.TxOuts {
				if output.Address == address {
					if _, ok := creatorTxs[tx.Id]; !ok {
						uTxOut := &UTxOut{tx.Id, index, output.Amount}
						if !isOnMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

func BalanceByAddress(address string, b *blockChain) int {
	txOuts := UTxOutsByAddress(address, b)
	var amount int
	for _, txOut := range txOuts {
		amount += txOut.Amount
	}
	return amount
}

func BlockChain() *blockChain {

	once.Do(func() {
		b = &blockChain{}
		// search for "checkpoint" on the db
		checkpoint := db.Checkpoint()
		if checkpoint == nil {
			b.AddBlock()
		} else {
			fmt.Printf("Restoring...")
			// restore b from bytes
			b.restore(checkpoint)
		}

	})

	//fmt.Printf("3NewestHash: %s\nHeight:%d", b.NewestHash, b.Height)
	return b
}
func Status(b *blockChain, rw http.ResponseWriter) {
	b.m.Lock()
	defer b.m.Unlock()
	utils.HandleErr(json.NewEncoder(rw).Encode(b))
}
func (b *blockChain) Replace(newBlocks []*Block) {
	b.m.Lock()
	defer b.m.Unlock()
	b.CurrentDifficulty = newBlocks[0].Difficulty
	b.Height = len(newBlocks)
	b.NewestHash = newBlocks[0].Hash
	persistBlockchain(b)
	db.EmptyBlocks()
	for _, block := range newBlocks {
		persistBlock(block)
	}
}
