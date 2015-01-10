package main

import (
	"fmt"
	btcd "github.com/eris-ltd/modules/btcd"
	"time"
)

func main() {
	b := btcd.NewBtcd()
	b.Init()
	b.Start()
	_, err := b.Get("newwallet", "passphrase")
	fmt.Println("get new wallet err:", err)
	f, err := b.Get("address")
	fmt.Println("address:", f, err)
	g, err := b.Get("accounts")
	fmt.Println("get accounts:", g)
	fmt.Println("get accounts err:", err)

	err = b.AutoCommit(true)
	fmt.Println("err on autocommit:", err)
	for {
		time.Sleep(time.Second)
	}
}
