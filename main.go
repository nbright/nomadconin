package main

import (
	"github.com/nbright/nomadcoin/cli"
	"github.com/nbright/nomadcoin/db"
)

func main() {
	//websocket, WS, 프로토콜
	db.DB()
	defer db.Close()
	cli.Start()
}
