package p2p

import (
	"encoding/json"
	"fmt"

	"github.com/nbright/nomadcoin/blockchain"
	"github.com/nbright/nomadcoin/utils"
)

type MessageKind int

const (
	MessageNewestBlock MessageKind = iota
	MessageAllBlocksRequest
	MessageAllBlocksResponse
	MessageNewBlockNotify
)

type Message struct {
	Kind    MessageKind
	Payload []byte
}

// kind와 payload로 JSON으로 인코딩
func makeMessage(kind MessageKind, payload interface{}) []byte {
	m := Message{
		Kind:    kind,
		Payload: utils.ToJSON(payload),
	}
	return utils.ToJSON(m)
}

// 3. 최신블록을 찾아 메시지로 만들어서 inbox로 보낸다.
func sendNewestBlock(p *peer) {
	fmt.Printf("Sending newest block to %s\n", p.key)
	b, err := blockchain.FindBlock(blockchain.BlockChain().NewestHash)
	utils.HandleErr(err)
	m := makeMessage(MessageNewestBlock, b)
	p.inbox <- m
}

func requestAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksRequest, nil)
	p.inbox <- m
}

func sendAllBlocks(p *peer) {
	m := makeMessage(MessageAllBlocksResponse, blockchain.Blocks(blockchain.BlockChain()))
	p.inbox <- m

}

func notifyNewBlock(b *blockchain.Block, p *peer) {
	m := makeMessage(MessageNewBlockNotify, b)
	p.inbox <- m
}

// 이 함수는 3000번에서 먼저 실행되는데 결국 4000번을 위해 실행
func handleMsg(m *Message, p *peer) {
	switch m.Kind {
	case MessageNewestBlock: // 처음에는 MessageNewestBlock이 port 3000번에 의해 먼저 실행됨.
		fmt.Printf("Received the newest block from %s\n", p.key)
		var payload blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		// 새로운 block 가져오기, 3000번은 4000번의 가장 새로운 블럭을 받음
		b, err := blockchain.FindBlock(blockchain.BlockChain().NewestHash)
		utils.HandleErr(err)
		//  4000              3000  :  4000번의 가장 새로운 block의 Height가 3000번의 가장 새로운 block보다 많다면 4000번의 Block들을 모두 요청함.
		if payload.Height >= b.Height {
			// request all the blocks from 4000
			fmt.Printf("Requesting all block from %s\n", p.key)
			requestAllBlocks(p)
		} else { // 아니면 4000번에 우리의 Block을 보냄.0
			//send 4000 our block
			sendNewestBlock(p)
		}
	case MessageAllBlocksRequest: // 4000번에 의해 실행
		fmt.Printf("%s wants all the blocks.\n", p.key)
		sendAllBlocks(p)
	case MessageAllBlocksResponse: // 3000번에 의해 실행
		fmt.Printf("Received all the blocks from %s\n", p.key)
		var payload []*blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		blockchain.BlockChain().Replace(payload)
	case MessageNewBlockNotify:

	}
	//fmt.Printf("Peer: %s, Sent a message with of: %d", p.key, m.Kind)
}
