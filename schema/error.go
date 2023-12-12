package schema

import (
	"encoding/json"
	"errors"
)

var (
	ERR_LARGER_DATA = errors.New("err_larger_data")

	ERR_TOKEN_NOT_EXIST    = errors.New("err_not_exist_token")
	ERR_BURN_FEE_NOT_EXIST = errors.New("err_not_exist_burn_fee")

	ERR_NOT_BUNDLE_TX = errors.New("err_not_bundle_tx")
	ERR_NOT_JSON_DATA = errors.New("err_not_json_data")

	ERR_EMAIL_CODE_EXPIRED   = errors.New("err_email_code_expired")
	ERR_REGISTER_SIG         = errors.New("err_register_sig")
	ERR_RP_ID_NOT_EXIST      = errors.New("err_rp_id_not_exist")
	ERR_ACC_TYPE_NOT_SUPPORT = errors.New("err_acc_type_not_support")
	ERR_SIGNER_INCORRECT     = errors.New("err_signer_incorrect")

	ERR_INVALID_ID                = errors.New("err_invalid_id")
	ERR_INVALID_OWNER             = errors.New("err_invalid_owner")
	ERR_INVALID_TX_VERSION        = errors.New("err_invalid_tx_version")
	ERR_INVALID_AMOUNT            = errors.New("err_invalid_amount")
	ERR_INVALID_FEE               = errors.New("err_invalid_fee")
	ERR_INVALID_TARGET_CHAIN_TYPE = errors.New("err_invalid_target_chain_type")
	ERR_INVALID_TX_MINT_HASH      = errors.New("err_invalid_mint_hash")
	ERR_INVALID_BUNDLE_DATA       = errors.New("err_invalid_bundle_data")
	ERR_INVALID_ACCOUNT_TYPE      = errors.New("err_invalid_account_type")
	ERR_INVALID_SIGNATURE         = errors.New("err_invalid_signature")

	ERR_NOT_FOUND_BUNDLE_SIG   = errors.New("err_not_found_bundle_sig")
	ERR_NOT_FOUND_BUNDLE_ITEMS = errors.New("err_not_found_bundle_items")

	ERR_BUNDLE_EXPIRED = errors.New("err_bundle_expired")
	ERR_BUNDLE_SALT    = errors.New("err_bundle_salt")
	ERR_BUNDLE_VERSION = errors.New("err_bundle_version")
)

type InternalErr struct {
	Index int    `json:"index"`
	Msg   string `json:"msg"`
}

// NewInternalErr if less than 0 (like -1), means not items error
func NewInternalErr(idx int, msg string) *InternalErr {
	return &InternalErr{
		Index: idx,
		Msg:   msg,
	}
}

func (e InternalErr) Error() string {
	jsErr, _ := json.Marshal(&e)
	return string(jsErr)
}
