package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/everFinance/goar/utils"
	"github.com/everVision/everpay-kits/schema"
)

func IDCheck(id string) (accountType, accID string, err error) {
	if common.IsHexAddress(id) {
		return schema.AccountTypeEVM, common.HexToAddress(id).String(), nil
	}

	if isArAddress(id) {
		return schema.AccountTypeAR, id, nil
	}

	if isEverId(id) {
		return schema.AccountTypeEverId, id, nil
	}

	return "", "", schema.ERR_INVALID_ID
}

func isArAddress(addr string) bool {
	if len(addr) != 43 {
		return false
	}
	_, err := utils.Base64Decode(addr)
	if err != nil {
		return false
	}

	return true
}

func isEverId(everId string) bool {
	// prefix must be ever
	if !strings.HasPrefix(everId, schema.EverIdPrefix) {
		return false
	}

	// length must 71
	if len(everId) != schema.EverIdLength {
		return false
	}

	// verify check sum
	// everId: eid36379ae9b4eb70465148bcc9b40e4a7d6c70e564c5de0f90ef017b60d01042aecc24
	fullBytes, err := hex.DecodeString(everId[3:])
	if err != nil {
		log.Error("hex.DecodeString(everId[3:])", "err", err)
		return false
	}

	length := len(fullBytes)
	sum := fullBytes[length-2:]
	bytesAddr := fullBytes[:length-2]
	trueSum := checkSum(bytesAddr)
	if bytes.Compare(sum, trueSum) != 0 {
		log.Error("check sum false")
		return false
	}
	return true
}

func checkSum(idBytes []byte) []byte {
	checkBytes := append([]byte(schema.EverIdPrefix), idBytes...) // eid+hash(email)
	hash := sha256.Sum256(checkBytes)
	return hash[:2]
}
