package blockchain

import (
	"errors"
	"time"

	"github.com/nbright/nomadcoin/utils"
	"github.com/nbright/nomadcoin/wallet"
)

const (
	minerReward int = 50
)

type mempool struct {
	Txs []*Tx
}

var Mempool *mempool = &mempool{}

type Tx struct {
	Id        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOuts"`
}

type TxIn struct {
	TxID  string
	Index int
	//서명
	Signature string `json:"signature"`
}
type TxOut struct {
	//address는 public key로 만들어지고 string으로 변환시킨 것.
	Address string `json:"address"`
	Amount  int
}

// unspentTxOut
type UTxOut struct {
	TxID   string
	Index  int
	Amount int
}

func (t *Tx) getId() {
	t.Id = utils.Hash(t)
}

// 서명을 우리가 서명한 Tx의 id를 갖는 Tx의 input에 저장.
func (t *Tx) sign() {
	for _, txIn := range t.TxIns {
		//우리가 가진 wallet의 private Key로 서명한다.
		txIn.Signature = wallet.Sign(t.Id, wallet.Wallet())
	}
}

func validate(tx *Tx) bool {
	valid := true
	for _, txIn := range tx.TxIns {
		// prevTx는 이 Tx input을 만든 이전의 Tx output의 TX를 말함.
		prevTx := FindTx(BlockChain(), txIn.TxID)
		if prevTx == nil {
			valid = false
			break
		}
		address := prevTx.TxOuts[txIn.Index].Address // 찾은 주소가 public Key임
		valid := wallet.Verify(txIn.Signature, tx.Id, address)
		if !valid {
			valid = false
			break
		}
	}
	return valid
}
func isOnMempool(uTxOut *UTxOut) bool {
	exists := false
Outer:
	for _, tx := range Mempool.Txs {
		for _, input := range tx.TxIns {
			if input.TxID == uTxOut.TxID && input.Index == uTxOut.Index {
				exists = true
				break Outer
			}
		}
	}
	return exists
}

func makeConinbaseTx(address string) *Tx {
	txIns := []*TxIn{
		{"", -1, "COINBASE"},
	}
	txOuts := []*TxOut{
		{address, minerReward},
	}
	tx := Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	return &tx

}

var ErrorNoMoney = errors.New("not enough 돈")
var ErrorNotValid = errors.New("Tx Invalid")

func makeTx(from, to string, amount int) (*Tx, error) {
	/* 순진한 코드
	if amount > BlockChain().BalanceByAddress(from) {
		return nil, errors.New("not enough money")
	}

	var txIns []*TxIn
	var txOuts []*TxOut
	total := 0
	oldTxOuts := BlockChain().TxOutsByAddress(from)
	for _, txOut := range oldTxOuts {
		if total > amount {
			break
		}
		txIn := &TxIn{txOut.Owner, txOut.Amount}
		txIns = append(txIns, txIn)
		total += txOut.Amount

	}
	change := total - amount
	if change != 0 {
		changeTxOut := &TxOut{from, change}
		txOuts = append(txOuts, changeTxOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)

	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		Txins:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	return tx, nil
	*/

	//보낼려고 하는 돈(Amount)가 from지갑주소의 돈보다 크면 못 보냄 (에러)

	if amount > BalanceByAddress(from, BlockChain()) {
		return nil, ErrorNoMoney
	}
	var txOuts []*TxOut
	var txIns []*TxIn
	total := 0
	uTxOuts := UTxOutsByAddress(from, BlockChain()) // from지갑주소에서 UnSpent Output 써버리지 않은 아웃풋
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := &TxIn{uTxOut.TxID, uTxOut.Index, from}
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	if change := total - amount; change != 0 {
		txOut := &TxOut{from, change}
		txOuts = append(txOuts, txOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)
	tx := &Tx{
		Id:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.getId()
	tx.sign()
	valid := validate(tx)
	if !valid {
		return nil, ErrorNoMoney
	}
	return tx, nil
}

func (m *mempool) AddTx(to string, amount int) error {
	tx, err := makeTx(wallet.Wallet().Address, to, amount)
	if err != nil {
		return err
	}
	m.Txs = append(m.Txs, tx)
	return nil
}

// 승인할 트랜잭션들 가져오기
func (m *mempool) TxToConfirm() []*Tx {
	coinbase := makeConinbaseTx(wallet.Wallet().Address)
	txs := m.Txs
	txs = append(txs, coinbase)
	m.Txs = nil
	return txs
}
