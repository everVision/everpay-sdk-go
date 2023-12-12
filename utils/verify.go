package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/everVision/everpay-kits/schema"
	"github.com/google/uuid"
)

func VerifyTransaction(tx schema.Transaction, payOwner string, chainID int) (tokenTx schema.TokenTransaction, pub []byte, err error) {
	// tx version
	if tx.Version != schema.TxVersionV1 {
		log.Error("invalid tx version", "version", tx.Version)
		err = schema.ERR_INVALID_TX_VERSION
		return
	}

	// owner verify
	if tx.ArOwner != payOwner {
		log.Error("invalid ar owner", "arOwer", tx.ArOwner)
		err = schema.ERR_INVALID_OWNER
		return
	}

	// data length more than 30k
	if len(tx.Data) > 30000 {
		log.Error("invalid data length", "len", len(tx.Data), "maxLen", 30000)
		err = schema.ERR_LARGER_DATA
		return
	}

	tokenTx, err = AsTokenTx(tx)
	if err != nil {
		log.Error("AsTokenTx", "err", err)
		return
	}

	acctype, accID, err := IDCheck(tx.From)
	if err != nil {
		log.Error("IDCheck", "err", err, "everHash", tx.HexHash())
		return
	}

	if tx.Action == schema.TxActionRegister {
		// fido account need verify email
		// notice: In the future there will be a lot of permission signing certifications for registering nicknames
		if acctype == schema.AccountTypeEverId {
			err = checkFidoEmail(tx.Nonce, tx.Data, tx.From, payOwner)
		}

		// there is no need to verify the register tx signature
		return
	}

	// Notice: ar owner in data early, need unmarshal and append to sig
	sig, err := FormatArSig(tx.Sig, tx.Data, acctype)
	if err != nil {
		log.Error("FormatArSig failed", "err", err, "everHash", tx.HexHash())
		return
	}
	nonce, err := strconv.ParseInt(tx.Nonce, 10, 64)
	if err != nil {
		return
	}
	pub, err = CompatVerify(nonce, acctype, accID, sig, tx.Hash(), tx.ArHash(), chainID)
	return
}

func VerifyBundleTransaction(data string, nonce int64, chainID int) (
	bundle *schema.BundleWithSigs,
	sigs map[string]string,
	pubs map[string][]byte,
	interErr *schema.InternalErr) {
	bundleData := &schema.BundleData{}
	if err := json.Unmarshal([]byte(data), &bundleData); err != nil {
		log.Error("can not unmarshal bundle data", "err", err)
		interErr = schema.NewInternalErr(-1, schema.ERR_INVALID_BUNDLE_DATA.Error())
		return
	}

	bundle = &bundleData.Bundle

	// items can not empty
	if len(bundleData.Bundle.Items) == 0 {
		interErr = schema.NewInternalErr(-1, schema.ERR_NOT_FOUND_BUNDLE_ITEMS.Error())
		return
	}

	// must be v1
	if strings.ToLower(bundleData.Bundle.Version) != schema.BundleTxVersionV1 {
		interErr = schema.NewInternalErr(-1, schema.ERR_BUNDLE_VERSION.Error())
		return
	}

	// salt must be uuid
	if _, err := uuid.Parse(bundleData.Bundle.Salt); err != nil {
		interErr = schema.NewInternalErr(-1, schema.ERR_BUNDLE_SALT.Error())
		return
	}

	// is expired: nonce is millisecond
	if bundle.Expiration < nonce/1000 {
		interErr = schema.NewInternalErr(-1, schema.ERR_BUNDLE_EXPIRED.Error())
		return
	}

	//  verify signature and get pubkeys
	sigs, pubs, interErr = VerifyBundleSigs(bundleData.Bundle, nonce, chainID)

	return
}

// return pubs: accID -> pubkey
func VerifyBundleSigs(stx schema.BundleWithSigs, nonce int64, chainID int) (sigs map[string]string, pubs map[string][]byte, err *schema.InternalErr) {
	// ensure every items through sig validation
	sigs = make(map[string]string) // key: accID, val: sig
	for id, sig := range stx.Sigs {
		_, accID, err := IDCheck(id)
		if err != nil {
			log.Error("account.IDCheck(id)", "err", err, "id", id)
			return nil, nil, schema.NewInternalErr(-1, err.Error())
		}
		if _, ok := sigs[accID]; ok {
			err = errors.New("exist same signer")
			return nil, nil, schema.NewInternalErr(-1, err.Error())
		}
		sigs[accID] = sig
	}

	pubs = map[string][]byte{} // key: accID, val: pubkey

	for idx, item := range stx.Items {
		acctype, accID, err := IDCheck(item.From)
		if err != nil {
			log.Error("account.IDCheck(id)", "err", err, "item.From", item.From)
			return nil, nil, schema.NewInternalErr(-1, err.Error())
		}

		// get sig
		sig, ok := sigs[accID]
		if !ok {
			log.Error("not found sig", "acc", accID)
			return nil, nil, schema.NewInternalErr(idx, schema.ERR_NOT_FOUND_BUNDLE_SIG.Error())
		}

		pub, vErr := CompatVerify(nonce, acctype, accID, sig, stx.Hash(), stx.ArHash(), chainID)
		if vErr == nil {
			pubs[accID] = pub
			continue
		}

		log.Error("invalid bundle sig", "err", vErr)
		return nil, nil, schema.NewInternalErr(idx, schema.ERR_INVALID_SIGNATURE.Error())
	}

	return
}

func CompatVerify(nonce int64, accType, accID string, sig string, hash, arHash []byte, chainID int) (public []byte, err error) {
	switch accType {
	case schema.AccountTypeEverId, schema.AccountTypeEVM:
		return Verify(accType, accID, sig, hash, chainID)

	case schema.AccountTypeAR:
		if nonce > 1701273600000 {
			return Verify(accType, accID, sig, arHash, chainID)
		}
		// old ar sig need to use hash or arHash
		public, err = Verify(accType, accID, sig, hash, chainID)
		if err != nil { // retry
			public, err = Verify(accType, accID, sig, arHash, chainID)
		}
	default:
		err = schema.ERR_ACC_TYPE_NOT_SUPPORT
	}
	return
}

func FormatArSig(sig, txData, accType string) (string, error) {
	if accType != schema.AccountTypeAR {
		return sig, nil
	}

	if strings.Contains(sig, ",") {
		return sig, nil
	}

	data := struct {
		ArOwner string `json:"arOwner"`
	}{}
	err := json.Unmarshal([]byte(txData), &data)
	if err != nil {
		return "", fmt.Errorf("unmarshal arOwner failed, err:%v", err)
	}

	return sig + "," + data.ArOwner, nil
}
