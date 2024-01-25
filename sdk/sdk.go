package sdk

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/everVision/everpay-kits/common"
	"github.com/everVision/everpay-kits/schema"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var log = common.NewLog("sdk")

type SDK struct {
	Info   schema.Info
	tokens map[string]schema.TokenInfo // tag -> TokenInfo

	signerType string // ecc, rsa
	signer     interface{}

	AccId string
	Cli   *Client

	lastNonce    int64 // last everTx used nonce
	sendTxLocker sync.Mutex
}

func New(signer interface{}, payUrl string) (*SDK, error) {
	signerType, signerAddr, err := reflectSigner(signer)
	if err != nil {
		return nil, err
	}

	sdk := &SDK{
		signer:       signer,
		signerType:   signerType,
		AccId:        signerAddr,
		Cli:          NewClient(payUrl),
		lastNonce:    time.Now().UnixNano() / 1000000,
		sendTxLocker: sync.Mutex{},
	}
	err = sdk.updatePayInfo()
	if err != nil {
		return nil, err
	}
	// sync info from everPay server every 10 mintue
	go sdk.runSyncInfo()
	return sdk, nil
}

func (s *SDK) runSyncInfo() {
	for {
		err := s.updatePayInfo()
		if err != nil {
			log.Error("can not get info from everpay", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		time.Sleep(10 * time.Minute)
	}
}

func (s *SDK) updatePayInfo() error {
	info, err := s.Cli.GetInfo()
	if err != nil {
		return err
	}

	tokens := make(map[string]schema.TokenInfo)
	for _, t := range info.TokenList {
		tokens[t.Tag] = t
	}
	s.tokens = tokens
	s.Info = info
	return nil
}

func (s *SDK) GetTokens() map[string]schema.TokenInfo {
	return s.tokens
}

func (s *SDK) SymbolToTagArr(symbol string) []string {
	tagArr := make([]string, 0)
	for tag, tok := range s.tokens {
		if strings.ToUpper(tok.Symbol) == strings.ToUpper(symbol) {
			tagArr = append(tagArr, tag)
		}
	}
	return tagArr
}

func (s *SDK) Transfer(tokenTag string, amount *big.Int, to, data string) (*schema.Transaction, error) {
	return s.sendTransfer(tokenTag, to, amount, data)
}

func (s *SDK) Withdraw(tokenTag string, amount *big.Int, chainType, to string) (*schema.Transaction, error) {
	return s.sendBurnTx(tokenTag, chainType, to, amount, "")
}

func (s *SDK) Deposit(tokenTag string, amount *big.Int, chainType, to, txData string) (*schema.Transaction, error) {
	return s.sendMintTx(tokenTag, chainType, to, amount, txData)
}

func (s *SDK) Burn(tokenTag string, amount *big.Int, chainType, to string) (*schema.Transaction, error) {
	return s.sendBurnTx(tokenTag, chainType, to, amount, "")
}

func (s *SDK) BurnToEverpay(tokenTag string, amount *big.Int) (*schema.Transaction, error) {
	chainType := schema.ChainTypeEverpay
	to := schema.ZeroAddress
	return s.sendBurnTx(tokenTag, chainType, to, amount, "")
}

func (s *SDK) Mint(tokenTag string, amount *big.Int, chainType, to, txData string) (*schema.Transaction, error) {
	return s.sendMintTx(tokenTag, chainType, to, amount, txData)
}

func (s *SDK) TransferTokenOwnerTx(tokenTag string, newOwner string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	return s.sendTx(tokenInfo, schema.TxActionTransferOwner, "0", newOwner, big.NewInt(0), "")
}

func (s *SDK) AddWhiteListTx(tokenTag string, whiteList []string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "whiteList", whiteList)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionAddWhiteList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) RemoveWhiteListTx(tokenTag string, whiteList []string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "whiteList", whiteList)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionRemoveWhiteList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) PauseWhiteListTx(tokenTag string, pause bool) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "pause", pause)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionPauseWhiteList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) AddBlackListTx(tokenTag string, blackList []string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "blackList", blackList)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionAddBlackList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) RemoveBlackListTx(tokenTag string, blackList []string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "blackList", blackList)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionRemoveBlackList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) PauseBlackListTx(tokenTag string, pause bool) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "pause", pause)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionPauseBlackList, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) PauseTokenTx(tokenTag string, pause bool) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	data, err := sjson.Set("", "pause", pause)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionPause, "0", s.AccId, big.NewInt(0), data)
}

func (s *SDK) Bundle(tokenTag string, to string, amount *big.Int, bundleWithSigs schema.BundleWithSigs) (*schema.Transaction, error) {
	bundle := schema.BundleData{
		Bundle: bundleWithSigs,
	}
	return s.sendBundle(tokenTag, to, amount, bundle)
}

func (s *SDK) sendTransfer(tokenTag string, receiver string, amount *big.Int, data string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	action := schema.TxActionTransfer
	fee := tokenInfo.TransferFee
	return s.sendTx(tokenInfo, action, fee, receiver, amount, data)
}

func (s *SDK) sendBurnTx(tokenTag string, targetChainType, receiver string, amount *big.Int, data string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	action := schema.TxActionBurn
	tFee, err := s.Cli.Fee(tokenTag)
	if err != nil {
		return nil, err
	}
	fee, ok := tFee.Fee.BurnFeeMap[targetChainType]
	if !ok {
		return nil, schema.ERR_BURN_FEE_NOT_EXIST
	}
	if data != "" && !gjson.Valid(data) {
		return nil, schema.ERR_NOT_JSON_DATA
	}

	// add targetChainType in data
	txData, err := sjson.Set(data, "targetChainType", targetChainType)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, action, fee, receiver, amount, txData)
}

func (s *SDK) sendMintTx(tokenTag string, targetChainType, receiver string, amount *big.Int, data string) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	if data != "" && !gjson.Valid(data) {
		return nil, schema.ERR_NOT_JSON_DATA
	}
	// add targetChainType in data
	txData, err := sjson.Set(data, "targetChainType", targetChainType)
	if err != nil {
		return nil, err
	}
	return s.sendTx(tokenInfo, schema.TxActionMint, "0", receiver, amount, txData)
}

func (s *SDK) sendBundle(tokenTag string, receiver string, amount *big.Int, bundle schema.BundleData) (*schema.Transaction, error) {
	tokenInfo, ok := s.tokens[tokenTag]
	if !ok {
		return nil, schema.ERR_TOKEN_NOT_EXIST
	}
	action := schema.TxActionBundle
	fee := tokenInfo.BundleFee

	data, err := json.Marshal(bundle)
	if err != nil {
		return nil, err
	}

	return s.sendTx(tokenInfo, action, fee, receiver, amount, string(data))
}

func (s *SDK) sendTx(tokenInfo schema.TokenInfo, action, fee, receiver string, amount *big.Int, data string) (*schema.Transaction, error) {
	s.sendTxLocker.Lock()
	defer s.sendTxLocker.Unlock()
	if amount == nil {
		amount = big.NewInt(0)
	}
	// assemble tx
	everTx := schema.Transaction{
		TokenSymbol:  tokenInfo.Symbol,
		Action:       action,
		From:         s.AccId,
		To:           receiver,
		Amount:       amount.String(),
		Fee:          fee,
		FeeRecipient: s.Info.FeeRecipient,
		Nonce:        fmt.Sprintf("%d", s.getNonce()),
		TokenID:      tokenInfo.ID,
		ChainType:    tokenInfo.ChainType,
		ChainID:      tokenInfo.ChainID,
		Data:         data,
		Version:      schema.TxVersionV1,
		Sig:          "",
	}

	sign, err := s.Sign(everTx.String())
	if err != nil {
		log.Error("Sign failed", "error", err)
		return &everTx, err
	}
	everTx.Sig = sign

	// submit to everpay server
	if err := s.Cli.SubmitTx(everTx); err != nil {
		log.Error("submit everTx", "error", err)
		return &everTx, err
	}

	return &everTx, nil
}

// about bundleTx

// GenBundle expiration: bundle tx expiration time(s)
func GenBundle(items []schema.BundleItem, expiration int64) schema.Bundle {
	return schema.Bundle{
		Items:      items,
		Expiration: expiration,
		Salt:       uuid.NewString(),
		Version:    schema.BundleTxVersionV1,
	}
}

func (s *SDK) SignBundleData(bundleTx schema.Bundle) (schema.BundleWithSigs, error) {
	sign, err := s.Sign(bundleTx.String())
	if err != nil {
		return schema.BundleWithSigs{}, err
	}
	return schema.BundleWithSigs{
		Bundle: bundleTx,
		Sigs: map[string]string{
			s.AccId: sign,
		},
	}, nil
}

func (s *SDK) getNonce() int64 {
	for {
		newNonce := time.Now().UnixNano() / 1000000
		if newNonce > s.lastNonce {
			s.lastNonce = newNonce
			return newNonce
		}
	}
}
