# tokucore – A Simple, Powerful Library for Bitcoin Apps.

[![Build Status](https://travis-ci.org/tokublock/tokucore.png)](https://travis-ci.org/tokublock/tokucore) [![Go Report Card](https://goreportcard.com/badge/github.com/tokublock/tokucore)](https://goreportcard.com/report/github.com/tokublock/tokucore) [![codecov.io](https://codecov.io/gh/tokublock/tokucore/graphs/badge.svg)](https://codecov.io/gh/tokublock/tokucore/branch/master) [![BSD License](http://img.shields.io/badge/license-BSD-blue.svg?style=flat)](LICENSE) <img src="http://segwit.co/static/public/images/logo.png" width="100">


## tokucore

*tokucore* is a simple Go (golang) library for creating and manipulating bitcoin data structures like creating keys and addresses (HD/BIP32/BIP39/SegWit) or parsing, creating and signing transactions, micropayment.

## Focus

* Simple and easy to use
* No external dependencies
* Full test coverage

## Tests

```
$ export GOPATH=`pwd`
$ go get -u github.com/tokublock/tokucore/xcore
$ cd src/github.com/tokublock/tokucore/
$ make test
```

## Examples

- [Generate a Random Address](examples/address_rand.go)
- [Generate a P2PKH Address](examples/address_p2pkh.go)
- [Generate a P2SH Address](examples/address_p2sh.go)
- [Generate a 2-of-3 P2SH MultiSig Address](examples/address_multisig.go)
- [Generate a P2WPKH SegWit Address](examples/address_p2wpkh.go)
- [Generate a P2WSH  SegWit Address](examples/address_p2wsh.go)
- [Create a P2PKH Transaction](examples/transaction_p2pkh.go)
- [Create a 2-to-3 P2SH MultiSig Transaction](examples/transaction_multisig.go)
- [Create a Transaction with an OP_RETURN Output](examples/transaction_opreturn.go)
- [Create a Transaction with Verify](examples/transaction_p2pkh.go)
- [Create a Transaction with P2WPKH Segwit Output](examples/transaction_p2wpkh.go)
- [Create a Transaction with P2WPKH SegWit Input](examples/transaction_p2wpkh.go)
- [Create a Transaction with P2WSH  SegWit Output](examples/transaction_p2wsh.go)
- [Create a Transaction with P2WSH  SegWit Input](examples/transaction_p2wsh.go)
- [HDWallet](examples/hdwallet.go)
- [Mnemonic](examples/bip39.go)

## Applications

- [JustDoBlockchain](https://justdoblockchain.com) - A website Learning Blockchain Demo by Demo.
- [ShaFish](https://shafish.com) - A web-based, trusted Bitcoin blockchain timestamping platform.

## Can I trust this code?
> Don't trust. Verify.

## License

tokucore is released under the BSD License.
