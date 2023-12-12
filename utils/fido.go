package utils

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/everFinance/goar/utils"
	"github.com/everVision/everpay-kits/schema"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/tidwall/gjson"
)

func GenUserId(eid string, chainID int) string {
	data := []byte(strings.ToLower(eid) + strconv.Itoa(chainID))
	hash := sha256.Sum256(data)
	userId := utils.Base64Encode(hash[:10])
	return userId
}

func GetWebAuthn(rpIdHash []byte) (*webauthn.WebAuthn, error) {
	localhostRpIdHash := sha256.Sum256([]byte(schema.LocalhostRpId))
	everpayioRpIdHash := sha256.Sum256([]byte(schema.EverpayRpId))
	switch string(rpIdHash) {
	case string(localhostRpIdHash[:]):
		return webAuthn(schema.LocalhostRpId)
	case string(everpayioRpIdHash[:]):
		return webAuthn(schema.EverpayRpId)
	default:
		return nil, schema.ERR_RP_ID_NOT_EXIST
	}
}

func webAuthn(rpId string) (*webauthn.WebAuthn, error) {
	return webauthn.New(&webauthn.Config{
		RPID:          rpId,
		RPDisplayName: "everpay",
		RPOrigins: []string{schema.EverpayOrg, schema.EverpayDevOrg, schema.BetaDevEverpayOrg, schema.BetaEverpayOrg,
			schema.LocalhostOrg},
		AttestationPreference: "",
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationRequired,
		},
		Debug:                false,
		EncodeUserIDAsString: false,
		Timeouts:             webauthn.TimeoutsConfig{},
	})
}

func decodeBase64(s string) (protocol.URLEncodedBase64, error) {
	// StdEncoding: the standard base64 encoded character set defined by RFC 4648, with the result padded with = so that the number of bytes is a multiple of 4
	// URLEncoding: another base64 encoded character set defined by RFC 4648, replacing '+' and '/' with '-' and '_'.
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")

	bs := &protocol.URLEncodedBase64{}
	err := bs.UnmarshalJSON([]byte(s))
	if err != nil {
		return nil, err
	}
	return *bs, nil
}

func checkFidoEmail(txNonce, txData string, everId string, payOwner string) error {
	jsData := gjson.Parse(txData)
	timestamp := jsData.Get("mailVerify").Get("timestamp").Int()

	nonce, err := strconv.ParseInt(txNonce, 10, 64)
	if err != nil {
		return err
	}
	timeRange := nonce/1000 - timestamp   // txNonce unit is ms
	if timeRange < 0 || timeRange > 300 { // timeRange 5 min
		log.Debug("acc register timeRange", "nonce", nonce, "timestamp", timestamp, "range", timeRange)
		return schema.ERR_EMAIL_CODE_EXPIRED
	}
	emailCode := jsData.Get("mailVerify").Get("code").String()
	codeSig := jsData.Get("mailVerify").Get("sig").String()

	msg := []byte(fmt.Sprintf("%d", timestamp) + emailCode + everId)
	sig, pubKey, addr, err := DecodeArSig(codeSig)
	if err != nil {
		return err
	}
	if addr != payOwner {
		return schema.ERR_REGISTER_SIG
	}
	return utils.Verify(accounts.TextHash(msg), pubKey, sig)
}
