package sdk

import (
	"testing"
	"time"

	"github.com/everVision/everpay-kits/schema"
	"github.com/stretchr/testify/assert"
)

var testClient *Client

func init() {
	testClient = NewClient("https://api-dev.everpay.io")
}

func TestGetInfo(t *testing.T) {
	info, err := testClient.GetInfo()
	assert.NoError(t, err)
	assert.Equal(t, "5", info.EthChainID)
}

func Test_IpWhite(t *testing.T) {
	// testClient.SetHeader("origin","https://auction.everpay.io")
	//
	// for {
	// 	go func() {
	// 		info, err := testClient.GetInfo()
	// 		assert.NoError(t, err)
	// 		t.Log(info.EthLocker)
	// 	}()
	// 	time.Sleep(10 * time.Millisecond)
	// }
}

func TestBalance(t *testing.T) {
	bal, err := testClient.Balance("ethereum-eth-0x0000000000000000000000000000000000000000", "0x2ca81e1253f9426c62Df68b39a22A377164eeC92")
	assert.NoError(t, err)
	assert.Equal(t, "0x2ca81e1253f9426c62Df68b39a22A377164eeC92", bal.AccId)
}

func TestBalances(t *testing.T) {
	bal, err := testClient.Balances("0x2ca81e1253f9426c62Df68b39a22A377164eeC92")
	assert.NoError(t, err)
	assert.Equal(t, "0x2ca81e1253f9426c62Df68b39a22A377164eeC92", bal.AccId)
}

func TestTxs(t *testing.T) {
	txs, err := testClient.Txs(1, "asc", 0, schema.TxOpts{})
	assert.NoError(t, err)
	assert.Equal(t, "ylP7Gu7vmImDRMX3K4K1TqtVT18ku0jC9gKS-XuTau8", txs.Txs[0].ID)
}

func TestTxByHash(t *testing.T) {
	tx, err := testClient.TxByHash("0x3b3b4caa8b9c1afbe3e683093815d07fe576a64a48b29c7c693922c76357cb7a")
	assert.NoError(t, err)
	assert.Equal(t, "TX-T3LLGF3pRkL42nymcZxx5ugFgRBEnLf2tmL_-tQ8", tx.Tx.ID)
}

func TestClient_BundleByHash(t *testing.T) {
	tx, bundle, status, err := testClient.BundleByHash("0xdf8c2a3ef9dc87d0a920bf4a3188928f22827d2220b0f6f481c555b45f2fbc4e")
	assert.NoError(t, err)
	assert.Equal(t, "0xdf8c2a3ef9dc87d0a920bf4a3188928f22827d2220b0f6f481c555b45f2fbc4e", tx.EverHash)
	assert.Equal(t, "0x72afe3c6537c3ed7eaf31db7faf5b0f4933561c3c34b630e3b7f530cb6ceede8", bundle.HashHex())
	assert.Equal(t, status.Status, "success")
	// assert.Equal(t, status.Index, 0)
	// assert.Equal(t, status.Msg, "err_insufficient_balance")
}

func TestClient_PendingTxs(t *testing.T) {
	txs, err := testClient.PendingTxs("")
	assert.NoError(t, err)
	t.Log(txs)
}

func TestClient_TokenFee(t *testing.T) {
	tag := "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee"
	fee, err := testClient.Fee(tag)
	assert.NoError(t, err)
	t.Log(fee)
}

func TestClient_SubscribeTxs_FilterToken(t *testing.T) {
	testClient = NewClient("https://api-dev.everpay.io")
	accid := "0x4002ED1a1410aF1b4930cF6c479ae373dEbD6223"
	sub := testClient.SubscribeTxs(schema.FilterQuery{
		Address:  accid,
		TokenTag: "arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xcc9141efa8c20c7df0778748255b1487957811be",
	})
	go func() {
		// for test
		time.Sleep(10 * time.Second)
		sub.Unsubscribe()
	}()

	for {
		select {
		case tx := <-sub.Subscribe():
			t.Log(tx.RawId, tx.EverHash)
		case <-sub.quit:
			return
		}
	}
}

func TestClient_SubscribeTxs_Cursor(t *testing.T) {
	testClient = NewClient("https://api-dev.everpay.io")
	accid := "0x4002ED1a1410aF1b4930cF6c479ae373dEbD6223"
	sub := testClient.SubscribeTxs(schema.FilterQuery{
		StartCursor: 155457,
		Address:     accid,
	})
	go func() {
		// for test
		time.Sleep(10 * time.Second)
		sub.Unsubscribe()
	}()

	for {
		select {
		case tx := <-sub.Subscribe():
			t.Log(tx.RawId, tx.EverHash)
		case <-sub.quit:
			return
		}
	}
}

func TestGetTokens(t *testing.T) {
	tokens, err := testClient.GetTokens()
	assert.NoError(t, err)
	t.Log(len(tokens), tokens)
}
