package main

import (
	"fmt"

	"github.com/nbright/nomadconin/blockchain"
)

func blockchainMain() {
	chain := blockchain.GetBlockChain()
	chain.AddBlock("Second Block")
	chain.AddBlock("Third Block")
	chain.AddBlock("Forth Block")

	for _, block := range chain.AllBlocks() {
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("PreHash: %s\n", block.PrevHash)
	}
}
