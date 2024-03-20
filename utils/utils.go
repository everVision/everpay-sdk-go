package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/everFinance/ethrpc"
	arTypes "github.com/everFinance/goar/types"
	"github.com/everVision/everpay-kits/common"
	"github.com/everVision/everpay-kits/schema"
)

var log = common.NewLog("utils")

func AsTokenTx(t schema.Transaction) (tokenTx schema.TokenTransaction, err error) {
	amount, ok := new(big.Int).SetString(t.Amount, 10)
	if !ok {
		log.Error("invalid amount", "amount", t.Amount)
		err = schema.ERR_INVALID_AMOUNT
		return
	}
	fee, ok := new(big.Int).SetString(t.Fee, 10)
	if !ok {
		log.Error("invalid fee", "fee", t.Fee)
		err = schema.ERR_INVALID_FEE
		return
	}

	_, from, err := IDCheck(t.From)
	if err != nil {
		return
	}
	_, to, err := IDCheck(t.To)
	if err != nil {
		return
	}
	feeRecipient := "" // if fee is 0, feeRecipient can be empty
	if t.Fee != "0" || t.FeeRecipient != "" {
		_, feeRecipient, err = IDCheck(t.FeeRecipient)
		if err != nil {
			return
		}
	}

	var targetChainType string
	if t.Action == schema.TxActionMint || t.Action == schema.TxActionBurn {
		targetChainType, err = GetTargetChainTypeFromData(t.Data, t.Action, t.ChainType)
		if err != nil {
			err = schema.ERR_INVALID_TARGET_CHAIN_TYPE
			return
		}
	}

	// bundle and set auto convert to transfer
	action := t.Action
	if t.Action == schema.TxActionBundle || t.Action == schema.TxActionSet {
		action = schema.TxActionTransfer
	}
	var mintHash string
	if action == schema.TxActionMint {
		mintHash, err = GetMintTargetTxHash(t.ChainType, t.Data, t.HexHash())
		if err != nil {
			log.Error("get mintHash failed", "err", err)
			err = schema.ERR_INVALID_TX_MINT_HASH
			return
		}
	}

	tokenTx = schema.TokenTransaction{
		MintHash:        mintHash,
		Action:          action,
		From:            from, // notice: Case Sensitive !!!
		To:              to,   // notice: Case Sensitive !!!
		Amount:          amount,
		Fee:             fee,
		FeeRecipient:    feeRecipient, // notice: Case Sensitive !!!
		TargetChainType: targetChainType,
		Data:            t.Data,
	}

	return
}

func GetTargetChainTypeFromData(txData, txAction, txChainType string) (string, error) {
	/*
		1. tns101 Token mint tx must have json txData
		2. when burn tx or chain type is 'everpay',
			if parsed targetChainType is "", then we think this is cross to native chain,so targetChainType == txChainType
	*/

	targetChain := struct{ TargetChainType string }{}
	err := json.Unmarshal([]byte(txData), &targetChain)
	// mint tx must have json txData and filter chain type is 'everpay' token
	if err != nil && txAction == schema.TxActionMint && txChainType != schema.ChainTypeEverpay {
		return "", err
	}

	targetChainType := targetChain.TargetChainType
	if targetChainType == "" {
		nativeChainType, err := GetEverToNativeChainType(txChainType)
		if err != nil {
			return "", err
		}
		targetChainType = nativeChainType
	}

	return targetChainType, nil
}

func GetEverToNativeChainType(everChainType string) (string, error) {
	switch everChainType {
	case schema.ChainTypeAos:
		return schema.OracleAosChainType, nil
	case schema.ChainTypeArweave, schema.ChainTypeCrossArEth:
		return schema.OracleArweaveChainType, nil
	case schema.ChainTypeEth:
		return schema.OracleEthChainType, nil
	case schema.ChainTypeMoonbeam, schema.ChainTypeMoonbase:
		return schema.OracleMoonChainType, nil
	case schema.ChainTypeCfx:
		return schema.OracleCfxChainType, nil
	case schema.ChainTypeBsc:
		return schema.OracleBscChainType, nil
	case schema.ChainTypePlaton:
		return schema.OraclePlatonChainType, nil
	case schema.ChainTypeEverpay:
		return schema.OracleEverpayChainType, nil
	default:
		return "", fmt.Errorf("not found this everChainType:%s", everChainType)
	}
}

func GetMintTargetTxHash(everTxChainType, everTxData string, everTxHash string) (targetChainTxHash string, err error) {
	targetChainType, err := GetTargetChainTypeFromData(everTxData, schema.TxActionMint, everTxChainType)
	if err != nil {
		return "", err
	}
	switch targetChainType { // more chain.
	case schema.OracleEthChainType, schema.OracleMoonChainType, schema.OracleCfxChainType, schema.OracleBscChainType, schema.OraclePlatonChainType:
		ethTx := ethrpc.Transaction{}
		if err := json.Unmarshal([]byte(everTxData), &ethTx); err != nil {
			log.Error("tx data unmarshal failed", "data", everTxData, "err", err)
			return "", err
		}
		return ethTx.Hash, nil
	case schema.OracleArweaveChainType:
		arTx := arTypes.Transaction{}
		if err = json.Unmarshal([]byte(everTxData), &arTx); err != nil {
			log.Error("tx data unmarshal failed, data", everTxData, "err", err)
			return
		}
		return arTx.ID, nil
	case schema.OracleEverpayChainType:
		return everTxHash, nil
	case schema.OracleAosChainType:
		// get data(items), first item
		items := struct {
			MainItem   arTypes.BundleItem `json:"mainItem"`
			PushedItem arTypes.BundleItem `json:"pushedItem"`
		}{}
		if err = json.Unmarshal([]byte(everTxData), &items); err != nil {
			log.Error("mint tx data unmarshal failed, data", everTxData, "err", err)
			return
		}
		if items.MainItem.Id == "" {
			log.Error("aos mintTx data incorrect")
			return "", errors.New("incorrect txData")
		}
		return items.MainItem.Id, nil
	default:
		err = fmt.Errorf("not support this targetChainType: %s", targetChainType)
		return
	}
}
