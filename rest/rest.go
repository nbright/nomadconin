package rest

import (
	"encoding/json"
	"fmt"

	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nbright/nomadcoin/blockchain"
	"github.com/nbright/nomadcoin/p2p"
	"github.com/nbright/nomadcoin/utils"
	"github.com/nbright/nomadcoin/wallet"
)

var port string

type url string

func (u url) MarshalText() ([]byte, error) {
	url := fmt.Sprintf("http://localhost%s%s", port, u)
	return []byte(url), nil
}

type addPayload struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type addPeerPayload struct {
	Address string `json:"address"`
	Port    string `json:"Port"`
}

type urlDescription struct {
	URL         url    `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

type blanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

type myWalletResponse struct {
	Address string `json:"address"`
}

//	func (u URLDescription) String() string {
//		return "Hello I'm URLDescription Print"
//	}
//
//	type addBlockBody struct {
//		Message string
//	}
type errorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

func documentation(rw http.ResponseWriter, r *http.Request) {
	data := []urlDescription{
		{
			URL:         url("/"),
			Method:      "GET",
			Description: "See Documentation",
		},
		{
			URL:         url("/status"),
			Method:      "GET",
			Description: "See the Status of the Blockchain",
		},
		{
			URL:         url("/blocks"),
			Method:      "GET",
			Description: "See All Blocks",
		},
		{
			URL:         url("/blocks"),
			Method:      "POST",
			Description: "Add A Block",
			Payload:     "data:string",
		},
		{
			URL:         url("/blocks/{hash}"),
			Method:      "GET",
			Description: "See A Block",
		},
		{
			URL:         url("/balance/{address}"),
			Method:      "GET",
			Description: "Get TxOuts for an Address0",
		},
		{
			URL:         url("/ws"),
			Method:      "GET",
			Description: "Upgrade to WebSockets",
		},
	}
	//rw.Header().Add("Content-Type", "application/json")
	// b, err := json.Marshal(data)
	// utils.HandleFunc(err)
	// fmt.Fprintf(rw, "%s", b)
	//위의 세줄 대신
	json.NewEncoder(rw).Encode(data)
}

func blocks(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(blockchain.Blocks(blockchain.BlockChain()))
	case "POST":
		// var addBlockBody addBlockBody
		// utils.HandleErr(json.NewDecoder(r.Body).Decode(&addBlockBody))
		newBlock := blockchain.BlockChain().AddBlock()

		rw.WriteHeader(http.StatusCreated)
	}
}
func block(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	block, err := blockchain.FindBlock(hash)
	encoder := json.NewEncoder(rw)
	if err == blockchain.ErrNotFound {
		encoder.Encode(errorResponse{fmt.Sprint(err)})
	} else {
		encoder.Encode(block)
	}
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(rw, r)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		next.ServeHTTP(rw, r)
	})
}

func status(rw http.ResponseWriter, r *http.Request) {
	blockchain.Status(blockchain.BlockChain(), rw)
}

func balance(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	total := r.URL.Query().Get("total")
	switch total {
	case "true":
		amount := blockchain.BalanceByAddress(address, blockchain.BlockChain())
		json.NewEncoder(rw).Encode(blanceResponse{address, amount})
	default:
		utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.UTxOutsByAddress(address, blockchain.BlockChain())))
	}

}

func mempool(rw http.ResponseWriter, r *http.Request) {
	utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.Mempool))
}
func transactions(rw http.ResponseWriter, r *http.Request) {
	var payload addPayload
	utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload))
	err := blockchain.Mempool.AddTx(payload.To, payload.Amount)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(errorResponse{err.Error()})
		return
	}
	rw.WriteHeader(http.StatusCreated)

}

func myWallet(rw http.ResponseWriter, r *http.Request) {
	address := wallet.Wallet().Address
	json.NewEncoder(rw).Encode(myWalletResponse{Address: address})

}
func peers(rw http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		var payload addPeerPayload
		utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload))
		p2p.AddPeer(payload.Address, payload.Port, port)
		rw.WriteHeader(http.StatusOK)
	case "GET":
		json.NewEncoder(rw).Encode(p2p.AllPeers(&p2p.Peers))
	}

	address := wallet.Wallet().Address
	json.NewEncoder(rw).Encode(myWalletResponse{Address: address})

}

func Start(aPort int) {
	router := mux.NewRouter()
	port = fmt.Sprintf(":%d", aPort)
	router.Use(jsonContentTypeMiddleware, loggerMiddleware)
	router.HandleFunc("/", documentation).Methods("GET")
	router.HandleFunc("/status", status)
	router.HandleFunc("/blocks", blocks).Methods("GET", "POST")

	//[a-f0-9]+ : 헥사데시멀 Reqular Expression
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET")
	router.HandleFunc("/balance/{address}", balance)
	router.HandleFunc("/mempool", mempool).Methods("GET")
	router.HandleFunc("/wallet", myWallet).Methods("GET")
	router.HandleFunc("/transactions", transactions).Methods("POST")
	router.HandleFunc("/ws", p2p.Upgrade).Methods("GET")
	router.HandleFunc("/peers", peers).Methods("GET", "POST")
	fmt.Printf("Listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
