package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"

	"os"

	"github.com/nbright/nomadcoin/utils"
)

const (
	fileName string = "nomadcoin.wallet"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

var w *wallet

func hasWalletFile() bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

func createPrivKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleErr(err)
	return privKey
}

func persistKey(key *ecdsa.PrivateKey) {
	bytes, err := x509.MarshalECPrivateKey(key)
	utils.HandleErr(err)
	err = os.WriteFile(fileName, bytes, 0644)
	utils.HandleErr(err)
}

// named return : 리턴할 변수를 미리 선언, 리턴시 변수 반환을 하지 않아도 됨.
// 아주 짧은 함수에서 사용할것을 권고
func restoreKey() (key *ecdsa.PrivateKey) {
	keyAsBytes, err := os.ReadFile(fileName)
	utils.HandleErr(err)
	key, err = x509.ParseECPrivateKey(keyAsBytes)
	utils.HandleErr(err)
	return
}

// 두개의 바이트 배열을 합쳐서 16진수로 인코딩
func encodeBigInts(a, b []byte) string {
	z := append(a, b...)
	return fmt.Sprintf("%x", z)
}

func aFromK(key *ecdsa.PrivateKey) string {
	return encodeBigInts(key.X.Bytes(), key.Y.Bytes())
}

// payLoad(Tx의 id)를 지갑의 privateKey 와 함께 넣어서 Sign 함.
func Sign(payload string, w *wallet) string {
	payloadAsB, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, payloadAsB)
	utils.HandleErr(err)
	return encodeBigInts(r.Bytes(), s.Bytes())
}

func restoreBigInts(signature string) (*big.Int, *big.Int, error) {
	bytes, err := hex.DecodeString(signature)
	utils.HandleErr(err)
	if err != nil {
		return nil, nil, err
	}
	firstHalfBytes := bytes[:len(bytes)/2]
	secondHalfBytes := bytes[len(bytes)/2:]
	bigA, bigB := big.Int{}, big.Int{}
	bigA.SetBytes(firstHalfBytes)
	bigB.SetBytes(secondHalfBytes)
	return &bigA, &bigB, nil
}

func Verify(signature, payload, address string) bool {

	r, s, err := restoreBigInts(signature)
	utils.HandleErr(err)
	x, y, err := restoreBigInts(address)
	utils.HandleErr(err)

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	payloadBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)

	ok := ecdsa.Verify(&publicKey, payloadBytes, r, s)
	fmt.Println(ok)
	return ok
}

func Wallet() *wallet {
	if w == nil {
		w = &wallet{}
		//has a wallet already?
		if hasWalletFile() {
			//yes -> 파일로 부터 지갑복구
			w.privateKey = restoreKey()
		} else {
			//no -> private Key 생성후 파일에 저장
			key := createPrivKey()
			persistKey(key)
			w.privateKey = key
		}
		w.Address = aFromK(w.privateKey)
	}
	return w
}

/*
	절차 /이론코드

1) we hash the msg
"i love you" -> hash(x) -> "hashed_message"
실제: msg 는 곧 Tx.ID 이다.

2) generate key pair
KeyPair (private Key, public Key) (save priv to a file)
실제: private Key는 wallet의 PrivateKey 이다.
      public Key는 이전 Tx 의 Address 이다.

3) sign the hash
("hashed_message" + private Key) -> "signature"


4) verify
("hashed_message" + "signature" + public Key) -> true / false

이론코드

	//2) generate key pair : create private key (Elliptic Curve Digital Signature Algorithem) y2=x3+ax+b
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	//x509 key 를 parse 하는 표준
	keyAsBytes, err := x509.MarshalECPrivateKey(privateKey)

	fmt.Printf("privateKey AsBytes: %x\n\n\n", keyAsBytes)

	utils.HandleErr(err)
	//1) we hash the msg

	fmt.Println("hashedMessage: ", hashedMessage)

	hashAsBytes, err := hex.DecodeString(hashedMessage)
	utils.HandleErr(err)

	//3) sign the hash : signature
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashAsBytes)

	signature := append(r.Bytes(), s.Bytes()...)
	fmt.Printf("signature: %x", signature)

	utils.HandleErr(err)

	//4) verify
	ok := ecdsa.Verify(&privateKey.PublicKey, hashAsBytes, r, s)
	fmt.Println(ok)

const (
	signature  string = "93cf8467f7b7379e303052e8834aa5a6c057e3a330769354f37645d62ff03bd9eda9ca698e9414b235741a6e43c2276e6fff43df02af0fcd53e57fd5984553e3"
	privateKey string = "307702010104208b83e5f622156466b69b0ac849b8d406e4623e90a2aa941c35f79d5ad021c29ea00a06082a8648ce3d030107a14403420004ed9444fe1e6a4177868680cbaea1bb680972eeb6135b5ed3823dba261ffc8f5fdd80d78b6564cda2d1a164bfe5e76db57265e889ed5a615ba4ad4e9c12496c4e"

	hashedMessage string = "1c5863cd55b5a4413fd59f054af57ba3c75c0698b3851d70f99b8de2d5c7338f"
)

func Start() {
	// 복구 과정
	privByte, err := hex.DecodeString(privateKey)
	utils.HandleErr(err)
	private, err := x509.ParseECPrivateKey(privByte)
	utils.HandleErr(err)
	fmt.Println(private)

	sigBytes, err := hex.DecodeString(signature)
	rBytes := sigBytes[:len(sigBytes)/2]
	sBytes := sigBytes[len(sigBytes)/2:]

	var bigR, bigS = big.Int{}, big.Int{}
	bigR.SetBytes(rBytes)
	bigS.SetBytes(sBytes)
	fmt.Println(bigR, bigS)

	hashBytes, err := hex.DecodeString(hashedMessage)

	ok := ecdsa.Verify(&private.PublicKey, hashBytes, &bigR, &bigS)
	fmt.Println(ok)
}
*/
