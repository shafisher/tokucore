// tokucore
//
// Copyright (c) 2018 TokuBlock
// BSD License

package xcore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tokublock/tokucore/network"
	"github.com/tokublock/tokucore/xcore/bip32"
	"github.com/tokublock/tokucore/xcrypto"
	"github.com/tokublock/tokucore/xerror"
)

func TestTransactionBuilderP2PKH(t *testing.T) {
	msg := []byte("666...satoshi")

	seed := []byte("this.is.bohu.seed.")
	bohuHDKey := bip32.NewHDKey(seed)
	bohuPrv := bohuHDKey.PrivateKey()
	bohuPub := bohuHDKey.PublicKey()
	bohu := NewPayToPubKeyHashAddress(bohuPub.Hash160())
	t.Logf("bohu.addr:%v", bohu.ToString(network.TestNet))

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	// Prepare the UTXOs.
	bohuCoin := NewCoinBuilder().AddOutput(
		"bde974a17f9ab1cfbbfb00bb4561e27156ebd65a4163ea0f014e9114d5b65556",
		1,
		6762017,
		"76a9145a927ddadc0ef3ae4501d0d9872b57c9584b9d8888ac",
	).ToCoins()[0]

	tx, err := NewTransactionBuilder().
		AddCoins(bohuCoin).
		AddKeys(bohuPrv).
		To(satoshi, 3000).
		Then().
		SetChange(bohu).
		SendFees(1000).
		Then().
		AddPushData(msg).
		Sign().
		BuildTransaction()
	assert.Nil(t, err)

	// Verify.
	err = tx.Verify()
	assert.Nil(t, err)

	t.Logf("%v", tx.ToString())
	t.Logf("txid:%v", tx.ID())

	assert.Equal(t, tx.BaseSize(), tx.Size())
	assert.Equal(t, tx.Vsize(), tx.Size())
	assert.Equal(t, "092ddeb0fa8205a06494f2cf83afda0377479c86065e60dea5ae347468b27361", tx.ID())

	t.Logf("basesize:%+v", tx.BaseSize())
	t.Logf("witnesssize:%+v", tx.WitnessSize())
	t.Logf("vsize:%+v", tx.Vsize())
	t.Logf("weight:%+v", tx.Weight())
	t.Logf("size:%+v", tx.Size())
	signedTx := tx.Serialize()
	t.Logf("actual.size:%v", len(signedTx))
	t.Logf("signed.tx:%x", signedTx)
}

func TestTransactionBuilderMultisigP2SH(t *testing.T) {
	seed := []byte("this.is.bohu.seed.")
	bohuHDKey := bip32.NewHDKey(seed)
	bohuPrv := bohuHDKey.PrivateKey()
	bohuPub := bohuHDKey.PublicKey()
	bohu := NewPayToPubKeyHashAddress(bohuPub.Hash160())
	t.Logf("bohu.addr:%v", bohu.ToString(network.TestNet))

	// A.
	seed = []byte("this.is.a.seed.")
	aHDKey := bip32.NewHDKey(seed)
	aPrv := aHDKey.PrivateKey()
	aPub := aHDKey.PublicKey().Serialize()

	// B.
	seed = []byte("this.is.b.seed.")
	bHDKey := bip32.NewHDKey(seed)
	bPub := bHDKey.PublicKey().Serialize()

	// C.
	seed = []byte("this.is.c.seed.")
	cHDKey := bip32.NewHDKey(seed)
	cPrv := cHDKey.PrivateKey()
	cPub := cHDKey.PublicKey().Serialize()

	// Redeem script.
	redeemScript := NewPayToMultiSigScript(2, aPub, bPub, cPub)
	redeem, _ := redeemScript.GetLockingScriptBytes()
	t.Logf("redeem.hex:%x", redeem)
	multi := NewPayToScriptHashAddress(xcrypto.Hash160(redeem))
	t.Logf("multi.addr:%v", multi.ToString(network.TestNet))

	// Funding.
	{
		bohuCoin := NewCoinBuilder().AddOutput(
			"092ddeb0fa8205a06494f2cf83afda0377479c86065e60dea5ae347468b27361",
			1,
			6758017,
			"76a9145a927ddadc0ef3ae4501d0d9872b57c9584b9d8888ac",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(bohuCoin).
			AddKeys(bohuPrv).
			To(multi, 4000).
			Then().
			SetChange(bohu).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)

		t.Logf("%v", tx.ToString())
		t.Logf("txid:%v", tx.ID())
		signedTx := tx.Serialize()
		t.Logf("funding.signed.tx:%x", signedTx)
		assert.Equal(t, "b2e955c95a6ee5752df1477a5936443ead0297ec697475ce6f356cdc6e2301a9", tx.ID())
	}

	// Spending.
	{
		multiCoin := NewCoinBuilder().AddOutput(
			"b2e955c95a6ee5752df1477a5936443ead0297ec697475ce6f356cdc6e2301a9",
			0,
			4000,
			"a914210a461ced66d7540ad2f4649b49dbed7c9fcc2887",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(multiCoin).
			AddKeys(aPrv, cPrv).
			SetRedeemScript(redeem).
			To(bohu, 1000).
			Then().
			SetChange(bohu).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)

		t.Logf("%v", tx.ToString())
		signedTx := tx.Serialize()
		t.Logf("txid:%v", tx.ID())
		t.Logf("spending.signed.tx:%x", signedTx)
		assert.Equal(t, "a28312ed5f5b5d164044f08f3a62e412aeb396043a1ec531c18994ff145ea793", tx.ID())
	}
}

func TestTransactionBuilderP2WPKH(t *testing.T) {
	seed := []byte("this.is.bohu.seed.")
	bohuHDKey := bip32.NewHDKey(seed)
	bohuPrv := bohuHDKey.PrivateKey()
	bohuPub := bohuHDKey.PublicKey()
	bohu := NewPayToPubKeyHashAddress(bohuPub.Hash160())
	t.Logf("bohu.addr:%v", bohu.ToString(network.TestNet))

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPrv := satoshiHDKey.PrivateKey()
	satoshiPubKey := satoshiHDKey.PublicKey()
	satoshi := NewPayToWitnessPubKeyHashAddress(satoshiPubKey.Hash160())
	t.Logf("satoshi.p2wpkh.addr:%v", satoshi.ToString(network.TestNet))

	// Funding.
	{
		bohuCoin := NewCoinBuilder().AddOutput(
			"f519a75190312039ddf885231205006b14f2e69f6e5b02314cb0e367b027fa86",
			1,
			127297408,
			"76a9145a927ddadc0ef3ae4501d0d9872b57c9584b9d8888ac",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(bohuCoin).
			AddKeys(bohuPrv).
			To(satoshi, 666666).
			Then().
			SetChange(bohu).
			SetRelayFeePerKb(20000).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)
		assert.Equal(t, "c37c3154ae611cfd9a57e684f0c12d51491d09060c643adc292565884e947b2b", tx.ID())

		t.Logf("fund:%v", tx.ToString())
		t.Logf("fund.txid:%v", tx.ID())
		t.Logf("fund.tx:%x", tx.Serialize())
		t.Logf("actualsize:%v", len(tx.Serialize()))
	}

	// Spending.
	{
		satoshiCoin := NewCoinBuilder().AddOutput(
			"c37c3154ae611cfd9a57e684f0c12d51491d09060c643adc292565884e947b2b",
			0,
			666666,
			"00148b7f2212ecc4384abcf1df3fc5783e9c2a24d5a5",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(satoshiCoin).
			AddKeys(satoshiPrv).
			To(bohu, 66666).
			Then().
			SetChange(satoshi).
			SetRelayFeePerKb(20000).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)

		t.Logf("spend:%v", tx.ToString())
		t.Logf("spend.txid:%v", tx.ID())
		t.Logf("spend.witnessid:%v", tx.WitnessID())
		t.Logf("spend.tx:%x", tx.Serialize())
		t.Logf("actualsize:%v", len(tx.Serialize()))
		assert.Equal(t, "80cd5fca2589cd97d3da1119214ed339d5284ce068e22f1eb9f32ee99a17d4bf", tx.ID())
	}
}

func TestTransactionBuilderP2WSH(t *testing.T) {
	seed := []byte("this.is.bohu.seed.")
	bohuHDKey := bip32.NewHDKey(seed)
	bohuPrv := bohuHDKey.PrivateKey()
	bohuPub := bohuHDKey.PublicKey()
	bohu := NewPayToPubKeyHashAddress(bohuPub.Hash160())
	t.Logf("bohu.addr:%v", bohu.ToString(network.TestNet))

	// A.
	seed = []byte("this.is.a.seed.")
	aHDKey := bip32.NewHDKey(seed)
	aPrv := aHDKey.PrivateKey()
	aPub := aHDKey.PublicKey().Serialize()

	// B.
	seed = []byte("this.is.b.seed.")
	bHDKey := bip32.NewHDKey(seed)
	bPub := bHDKey.PublicKey().Serialize()

	// C.
	seed = []byte("this.is.c.seed.")
	cHDKey := bip32.NewHDKey(seed)
	cPrv := cHDKey.PrivateKey()
	cPub := cHDKey.PublicKey().Serialize()

	// Redeem script.
	redeemScript := NewPayToMultiSigScript(2, aPub, bPub, cPub)
	redeem, _ := redeemScript.GetLockingScriptBytes()
	t.Logf("redeem.hex:%x", redeem)
	multi := NewPayToWitnessScriptHashAddress(xcrypto.Sha256(redeem))
	t.Logf("multi.addr:%v", multi.ToString(network.TestNet))
	assert.Equal(t, "tb1qrrf2qzw8stxkwhurtamy7wkl3a24vhgu0l3gcf66a8hl5dk9napqtap6rf", multi.ToString(network.TestNet))

	// Funding.
	{
		bohuCoin := NewCoinBuilder().AddOutput(
			"b2e955c95a6ee5752df1477a5936443ead0297ec697475ce6f356cdc6e2301a9",
			1,
			6753017,
			"76a9145a927ddadc0ef3ae4501d0d9872b57c9584b9d8888ac",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(bohuCoin).
			AddKeys(bohuPrv).
			To(multi, 4000).
			Then().
			SetChange(bohu).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)

		assert.Equal(t, "02f96826dbd8bfec2e88603d110dfef1872809debfd84c12188ab94097da3998", tx.ID())

		t.Logf("%v", tx.ToString())
		t.Logf("txid:%v", tx.ID())
		signedTx := tx.Serialize()
		t.Logf("funding.signed.tx:%x", signedTx)
	}

	// Spending.
	{
		multiCoin := NewCoinBuilder().AddOutput(
			"02f96826dbd8bfec2e88603d110dfef1872809debfd84c12188ab94097da3998",
			0,
			4000,
			"002018d2a009c782cd675f835f764f3adf8f55565d1c7fe28c275ae9effa36c59f42",
		).ToCoins()[0]

		tx, err := NewTransactionBuilder().
			AddCoins(multiCoin).
			AddKeys(aPrv, cPrv).
			SetRedeemScript(redeem).
			To(bohu, 1000).
			Then().
			SetChange(bohu).
			Then().
			Sign().
			BuildTransaction()
		assert.Nil(t, err)

		// Verify.
		err = tx.Verify()
		assert.Nil(t, err)

		t.Logf("%v", tx.ToString())
		signedTx := tx.Serialize()
		t.Logf("txid:%v", tx.ID())
		t.Logf("spending.signed.tx:%x", signedTx)
		assert.Equal(t, "70eaf6275e59e780b933d88ea87b0d1f3135ea2ecb6add971f975155ec80d918", tx.ID())
	}
}

func TestTransactionBuilderWithUncompressedPubKey(t *testing.T) {
	seed := []byte("this.is.bohu.seed.")
	bohuHDKey := bip32.NewHDKey(seed)
	bohuPrv := bohuHDKey.PrivateKey()
	bohuPub := bohuPrv.PubKey()

	// Uncompressed pubkey.
	pubHash := xcrypto.Hash160(bohuPub.SerializeUncompressed())
	script := NewPayToPubKeyHashScript(pubHash)
	bohu := script.GetAddress()
	locking, err := script.GetRawLockingScriptBytes()
	assert.Nil(t, err)

	t.Logf("bohu.addr:%v", bohu.ToString(network.TestNet))

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	// Prepare the UTXOs.
	bohuCoin := NewCoinBuilder().AddOutput(
		"5af1520f1d3e818fca695c2a903baa4a7eec4954f0b35aa01be1f2c1d2cfd802",
		0,
		129990000,
		fmt.Sprintf("%x", locking),
	).ToCoins()[0]

	tx, err := NewTransactionBuilder().
		AddCoins(bohuCoin).
		AddKeys(bohuPrv).
		To(satoshi, 666666).
		SetPubKeyUncompressed().
		Then().
		SetChange(bohu).
		SendFees(10000).
		Then().
		Sign().
		BuildTransaction()
	assert.Nil(t, err)

	// Verify.
	err = tx.Verify()
	assert.Nil(t, err)
}

func TestTransactionBuilderHybrid(t *testing.T) {
	// Alice.
	seed := []byte("this.is.alice.seed.")
	aliceHDKey := bip32.NewHDKey(seed)
	alicePrv := aliceHDKey.PrivateKey()
	alicePub := aliceHDKey.PublicKey()
	alice := NewPayToPubKeyHashAddress(alicePub.Hash160())
	aliceCoin := MockP2PKHCoin(aliceHDKey)

	// Bob.
	seed = []byte("this.is.bob.seed.")
	bobHDKey := bip32.NewHDKey(seed)
	bobPrv := bobHDKey.PrivateKey()
	bobPub := bobHDKey.PublicKey()
	bobCoin := MockP2PKHCoin(bobHDKey)

	// Alice and bob.
	redeem, _ := NewPayToMultiSigScript(2, alicePub.Serialize(), bobPub.Serialize()).GetLockingScriptBytes()
	aliceBobCoin := MockP2SHCoin(aliceHDKey, bobHDKey, redeem)

	// AD.
	pushData := []byte("this.is.pushdata")

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	tx, err := NewTransactionBuilder().
		AddCoins(aliceCoin).
		AddKeys(alicePrv).
		To(satoshi, 10000).
		Then().
		AddCoins(bobCoin).
		AddKeys(bobPrv).
		To(satoshi, 9000).
		Then().
		AddCoins(aliceBobCoin).
		AddKeys(alicePrv, bobPrv).
		SetRedeemScript(redeem).
		To(satoshi, 20000).
		Then().
		SetChange(alice).
		SendFees(1000).
		Then().
		AddPushData(pushData).
		Sign().
		BuildTransaction()
	assert.Nil(t, err)
	signedTx := tx.Serialize()
	err = tx.Verify()
	assert.Nil(t, err)
	t.Logf("signed.hex:%x", signedTx)
	t.Logf("signed.string:%v", tx.ToString())
}

func TestTransactionBuilderFees(t *testing.T) {
	// Alice.
	seed := []byte("this.is.alice.seed.")
	aliceHDKey := bip32.NewHDKey(seed)
	alicePrv := aliceHDKey.PrivateKey()
	alicePub := aliceHDKey.PublicKey()
	alice := NewPayToPubKeyHashAddress(alicePub.Hash160())
	aliceCoin := MockP2PKHCoin(aliceHDKey)

	// Bob.
	seed = []byte("this.is.bob.seed.")
	bobHDKey := bip32.NewHDKey(seed)
	bobPrv := bobHDKey.PrivateKey()
	bobPub := bobHDKey.PublicKey()
	bobCoin := MockP2PKHCoin(bobHDKey)

	// Alice and bob.
	redeem, _ := NewPayToMultiSigScript(2, alicePub.Serialize(), bobPub.Serialize()).GetLockingScriptBytes()
	aliceBobCoin := MockP2SHCoin(aliceHDKey, bobHDKey, redeem)

	// AD.
	pushData := []byte("this.is.pushdata")

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	tx, err := NewTransactionBuilder().
		AddCoins(aliceCoin).
		AddKeys(alicePrv).
		To(satoshi, 10000).
		Then().
		AddCoins(bobCoin).
		AddKeys(bobPrv).
		To(satoshi, 9000).
		Then().
		AddCoins(aliceBobCoin).
		AddKeys(alicePrv, bobPrv).
		SetRedeemScript(redeem).
		To(satoshi, 20000).
		Then().
		SetChange(alice).
		SetRelayFeePerKb(100).
		Then().
		AddPushData(pushData).
		Sign().
		BuildTransaction()
	assert.Nil(t, err)
	signedTx := tx.Serialize()
	t.Logf("actual.size:%v", len(signedTx))
}

func TestTransactionBuilderError(t *testing.T) {
	// Alice.
	seed := []byte("this.is.alice.seed.")
	aliceHDKey := bip32.NewHDKey(seed)
	alicePrv := aliceHDKey.PrivateKey()
	alicePub := aliceHDKey.PublicKey()
	alice := NewPayToPubKeyHashAddress(alicePub.Hash160())
	aliceCoin := MockP2PKHCoin(aliceHDKey)

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	tests := []struct {
		name string
		fn   func() error
		err  error
	}{
		{
			name: "builder.from.nil",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddKeys(alicePrv).
					To(satoshi, 10000).
					Then().
					SetChange(alice).
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_FROM_EMPTY, 0),
		},
		{
			name: "builder.sendto.nil",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					AddKeys(alicePrv).
					Then().
					SetChange(alice).
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_SENDTO_EMPTY, 0),
		},
		{
			name: "builder.change.nil",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					AddKeys(alicePrv).
					To(satoshi, 1000).
					Then().
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_CHANGETO_EMPTY),
		},
		{
			name: "builder.fee.not.enough",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					AddKeys(alicePrv).
					To(satoshi, 10000).
					Then().
					SetChange(alice).
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_MIN_FEE_NOT_ENOUGH, 1000, 0),
		},
		{
			name: "builder.totalout.more.than.totalin",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					AddKeys(alicePrv).
					To(satoshi, 1000000).
					Then().
					SetChange(alice).
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_AMOUNT_NOT_ENOUGH_ERROR, 1000000, 10000),
		},
		{
			name: "builder.keys.nil",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					To(satoshi, 1000).
					Then().
					SetChange(alice).
					SendFees(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_SIGN_KEY_EMPTY, 0),
		},
		{
			name: "builder.fee.high",
			fn: func() error {
				_, err := NewTransactionBuilder().
					AddCoins(aliceCoin).
					To(satoshi, 1000).
					Then().
					SetChange(alice).
					SetMaxFees(10).
					SetRelayFeePerKb(1000).
					Then().
					Sign().
					BuildTransaction()
				return err
			},
			err: xerror.NewError(Errors, ER_TRANSACTION_BUILDER_FEE_TOO_HIGH, 192, 10),
		},
	}
	for _, test := range tests {
		err := test.fn()
		assert.Equal(t, test.err.Error(), err.Error())
	}
}

func BenchmarkTransactionBuilder(b *testing.B) {
	// Alice.
	seed := []byte("this.is.alice.seed.")
	aliceHDKey := bip32.NewHDKey(seed)
	alicePrv := aliceHDKey.PrivateKey()
	alicePub := aliceHDKey.PublicKey()
	alice := NewPayToPubKeyHashAddress(alicePub.Hash160())
	aliceCoin := MockP2PKHCoin(aliceHDKey)

	// Bob.
	seed = []byte("this.is.bob.seed.")
	bobHDKey := bip32.NewHDKey(seed)
	bobPrv := bobHDKey.PrivateKey()
	bobCoin := MockP2PKHCoin(bobHDKey)

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	for n := 0; n < b.N; n++ {
		_, err := NewTransactionBuilder().
			AddCoins(aliceCoin).
			AddKeys(alicePrv).
			To(satoshi, 5000).
			Then().
			AddCoins(bobCoin).
			AddKeys(bobPrv).
			To(satoshi, 5000).
			Then().
			SetChange(alice).
			Then().
			BuildTransaction()
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkTransactionBuilderSigned(b *testing.B) {
	// Alice.
	seed := []byte("this.is.alice.seed.")
	aliceHDKey := bip32.NewHDKey(seed)
	alicePrv := aliceHDKey.PrivateKey()
	alicePub := aliceHDKey.PublicKey()
	alice := NewPayToPubKeyHashAddress(alicePub.Hash160())
	aliceCoin := MockP2PKHCoin(aliceHDKey)

	// Bob.
	seed = []byte("this.is.bob.seed.")
	bobHDKey := bip32.NewHDKey(seed)
	bobPrv := bobHDKey.PrivateKey()
	bobCoin := MockP2PKHCoin(bobHDKey)

	// Satoshi.
	seed = []byte("this.is.satoshi.seed.")
	satoshiHDKey := bip32.NewHDKey(seed)
	satoshiPub := satoshiHDKey.PublicKey()
	satoshi := NewPayToPubKeyHashAddress(satoshiPub.Hash160())

	for n := 0; n < b.N; n++ {
		_, err := NewTransactionBuilder().
			AddCoins(aliceCoin).
			AddKeys(alicePrv).
			To(satoshi, 5000).
			Then().
			AddCoins(bobCoin).
			AddKeys(bobPrv).
			To(satoshi, 5000).
			Then().
			SetChange(alice).
			Then().
			Sign().
			BuildTransaction()
		if err != nil {
			panic(err)
		}
	}
}
