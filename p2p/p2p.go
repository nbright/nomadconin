package p2p

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nbright/nomadcoin/blockchain"
	"github.com/nbright/nomadcoin/utils"
)

var upgrader = websocket.Upgrader{}

// 2
func Upgrade(rw http.ResponseWriter, r *http.Request) {
	openPort := r.URL.Query().Get("openPort")
	ip := utils.Splitter(r.RemoteAddr, ":", 0)
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return openPort != "" && ip != ""
	}
	// Port 3000 에서 4000으로 메시지를 보낼수 있는 문이야.
	// 3000에 이 컨넥션이 만들어짐.
	// Port 3000은 4000포트로 부터 요청받아 업그레이드 할 것임.
	fmt.Printf("%s wants an upgrade\n", openPort)
	conn, err := upgrader.Upgrade(rw, r, nil)
	utils.HandleErr(err)
	peer := initPeer(conn, ip, openPort)
	time.Sleep(10 * time.Second)
	peer.inbox <- []byte("Hello from 3000!")
	//conn.WriteMessage(websocket.TextMessage, []byte("Hello from Port 3000!"))

}

// 1
func AddPeer(address, port, openPort string) {
	// from :4000 에서 :3000으로 보낼수 있는 Conn 을 만듬. 4000에 이 컨넥션이 만들어짐.
	// Port 4000은 Port 3000으로부터 업그레이드를 요청하고 있음.
	fmt.Printf("%s wants to connect to port %s\n", openPort, port)
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws?openPort=%s", address, port, openPort[1:]), nil)
	utils.HandleErr(err)
	p := initPeer(conn, address, port)
	sendNewestBlock(p)
	//time.Sleep(5 * time.Second)
	//peer.inbox <- []byte("Hello from 4000!")
	//conn.WriteMessage(websocket.TextMessage, []byte("Hello from Port 4000!"))
}

func BroadcastNewBlock(b *blockchain.Block) {
	for _, p := range Peers.v {
		notifyNewBlock(b, p)
	}
}

/** 아주 중요,서버에서  메시지 읽어서, 보내기
var conns []*websocket.Conn
func Upgrade(rw http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(rw, r, nil)
	conns.append(conns, conn)
	utils.HandleErr(err)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {

			break
		}
		for _, aConn := range conns {
			if aConn != conn {
				aConn.WriteMessage(websocket.TextMessage, p)
			}
		}

	}

}
*/
