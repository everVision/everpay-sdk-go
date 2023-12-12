package schema

import "errors"

var (
	ERR_INVALID_ID         = errors.New("err_invalid_id")
	ERR_TOKEN_NOT_EXIST    = errors.New("err_not_exist_token")
	ERR_BURN_FEE_NOT_EXIST = errors.New("err_not_exist_burn_fee")
	ERR_NOT_BUNDLE_TX      = errors.New("err_not_bundle_tx")
	ERR_NOT_JSON_DATA      = errors.New("err_not_json_data")
)
