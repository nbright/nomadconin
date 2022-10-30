package main

import (
	"fmt"
	"time"
)

func send(c chan<- int) { // 보내기(채널에 쓰기)전용 채널 send only channel
	for i := range [10]int{} {
		fmt.Printf(">> sending %d <<\n", i)
		c <- i // blocking 임. 누군가가 받을 준비가 되어야
		fmt.Printf(">>sent %d <<\n", i)
	}
	close(c)
}

func receive(c <-chan int) { // 받기(채널에서 읽기)전용 채널 receive only channel
	for {
		time.Sleep(5 * time.Second)
		a, ok := <-c //blocking 기다리고 있음. block operation
		if !ok {
			fmt.Printf("we are done")
			break
		}
		fmt.Printf("|| received %d ||\n", a)
	}
}

func Channelmain() {
	// //go explorer.Start(3000)
	// //rest.Start(4000)
	// //fmt.Println(os.Args[2:])
	// //db.DB()
	// // blockchain.BlockChain().AddBlock("First")
	// // blockchain.BlockChain().AddBlock("Second")
	// // blockchain.BlockChain().AddBlock("Third")
	// defer db.Close()
	// cli.Start()
	// //wallet.Wallet()

	//숫자 5개가 올때마다 block 하는 Buffered channel 임. size=5 큐와 같은원리
	// channel의 크기가 5개라서 5개의 메시지를 보내면 꽉차서 Receive 가 1개 일어나면 다시 하나 보낼수 있음.
	c := make(chan int, 10)
	go send(c)
	receive(c)

}
