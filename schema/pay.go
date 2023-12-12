package schema

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	BundleTxVersionV1 = "v1"
	TxActionBundle    = "bundle"
	TxActionSet       = "set"
	TxActionRegister  = "register"
)

type Transaction struct {
	TokenSymbol  string `json:"tokenSymbol"`
	Action       string `json:"action"`
	From         string `json:"from"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
	Fee          string `json:"fee"`
	FeeRecipient string `json:"feeRecipient"`
	Nonce        string `json:"nonce"`
	TokenID      string `json:"tokenID"`
	ChainType    string `json:"chainType"`
	ChainID      string `json:"chainID"`
	Data         string `json:"data"`
	Version      string `json:"version"`
	Sig          string `json:"sig"`

	ArOwner     string `json:"-"`
	ArTxID      string `json:"-"`
	ArTimestamp int64  `json:"-"`
}

func (t *Transaction) String() string {
	return "tokenSymbol:" + t.TokenSymbol + "\n" +
		"action:" + t.Action + "\n" +
		"from:" + t.From + "\n" +
		"to:" + t.To + "\n" +
		"amount:" + t.Amount + "\n" +
		"fee:" + t.Fee + "\n" +
		"feeRecipient:" + t.FeeRecipient + "\n" +
		"nonce:" + t.Nonce + "\n" +
		"tokenID:" + t.TokenID + "\n" +
		"chainType:" + t.ChainType + "\n" +
		"chainID:" + t.ChainID + "\n" +
		"data:" + t.Data + "\n" +
		"version:" + t.Version
}

// Tag is the unique identifier of token
func (t *Transaction) Tag() string {
	return tag(t.ChainType, t.TokenSymbol, t.TokenID)
}

func (t *Transaction) Hash() []byte {
	return accounts.TextHash([]byte(t.String()))
}

func (t *Transaction) HexHash() string {
	return hexutil.Encode(t.Hash())
}

func (t *Transaction) ArHash() []byte {
	msg := sha256.Sum256([]byte(t.String()))
	return msg[:]
}

type BalanceTreeHash struct {
	RootHash string `json:"rootHash"`
	EverHash string `json:"everHash"`
}

type InternalStatus struct {
	Status string `json:"status"` // "success" or "failed"
	*InternalErr
}

func (i InternalStatus) Marshal() string {
	js, _ := json.Marshal(i)
	return string(js)
}

type BundleItem struct {
	Tag     string `json:"tag"` // token tag
	ChainID string `json:"chainID"`
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  string `json:"amount"`
}

type Bundle struct {
	Items      []BundleItem `json:"items"`
	Expiration int64        `json:"expiration"` // second
	Salt       string       `json:"salt"`       // uuid
	Version    string       `json:"version"`
}

type BundleWithSigs struct {
	Bundle
	Sigs map[string]string `json:"sigs"` // accID -> sig
}

type BundleData struct {
	Bundle BundleWithSigs `json:"bundle"`
}

func (s *Bundle) String() string {
	by, _ := json.Marshal(s)
	return string(by)
}

func (s *Bundle) Hash() []byte {
	return accounts.TextHash([]byte(s.String()))
}

func (s *Bundle) HashHex() string {
	return hexutil.Encode(s.Hash())
}

func (s *Bundle) ArHash() []byte {
	msg := sha256.Sum256([]byte(s.String()))
	return msg[:]
}
