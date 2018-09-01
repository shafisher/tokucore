// tokucore
//
// Copyright (c) 2018 TokuBlock
// BSD License

package main

import (
	"fmt"

	"github.com/tokublock/tokucore/xcore"
)

// Bitcoin HD wallet demo.
func main() {
	seed := []byte("bitcoin blockchain tokublock sandbox")
	hdkey := xcore.NewHDKey(seed)

	// Master Private Key.
	masterprv := hdkey.ToString(xcore.TestNet)
	fmt.Printf("master.prvkey:%v\n", masterprv)

	// bitcoin  path: m/44'/0'/0'/0
	{
		for i := 0; i < 2; i++ {
			path := fmt.Sprintf("m/44'/0'/0'/%d", i)
			prvkey, err := hdkey.DeriveByPath(path)
			if err != nil {
				panic(err)
			}
			fmt.Printf("btc.chain:%v\n", path)
			fmt.Printf("\tchild.prvkey:%v\n", prvkey.ToString(xcore.TestNet))

			pubkey := prvkey.HDPublicKey()
			fmt.Printf("\tchild.pubkey:%v\n", pubkey.ToString(xcore.TestNet))
		}
	}

	// ethereum path: m/44'/60'/0'/0
	{
		for i := 0; i < 2; i++ {
			path := fmt.Sprintf("m/44'/60'/0'/%d", i)
			prvkey, err := hdkey.DeriveByPath(path)
			if err != nil {
				panic(err)
			}
			fmt.Printf("eth.chain:%v\n", path)
			fmt.Printf("\tchild.prvkey:%v\n", prvkey.ToString(xcore.TestNet))

			pubkey := prvkey.HDPublicKey()
			fmt.Printf("\tchild.pubkey:%v\n", pubkey.ToString(xcore.TestNet))
		}
	}
}
