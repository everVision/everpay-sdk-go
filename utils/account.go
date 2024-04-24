package utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/everFinance/goar/utils"
	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/schema"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func Verify(accType, accID string, sig string, hash []byte, chainID int) (public []byte, err error) {
	switch accType {
	case schema.AccountTypeEVM:
		var addr common.Address
		public, addr, err = goether.Ecrecover(hash, common.FromHex(sig))
		if err != nil {
			return nil, err
		}
		if strings.ToLower(accID) != strings.ToLower(addr.String()) {
			err = schema.ERR_SIGNER_INCORRECT
			return
		}
	case schema.AccountTypeAR:
		sig1, pubkey, addr, err := DecodeArSig(sig)
		if err != nil {
			return nil, err
		}
		if addr != accID {
			return nil, schema.ERR_SIGNER_INCORRECT
		}

		if err = utils.Verify(hash, pubkey, sig1); err != nil {
			return nil, err
		}
		public = pubkey.N.Bytes()
	case schema.AccountTypeEverId:
		// txSig = sig + "," + public + "," + publicType
		public, err = verifyEverIdSig(accID, sig, hash, chainID)
		if err != nil {
			return
		}
	default:
		return nil, schema.ERR_INVALID_ACCOUNT_TYPE
	}
	return public, nil
}

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

func verifyEverIdSig(accID, txSig string, hash []byte, chainID int) ([]byte, error) {
	// txSig = sig + "," + public + "," + publicType
	// eccSig = eccSig,base64(publicBy),"ECDSA"
	// arSig = arSig,base64(publicBy),"RSA"
	// fidoSig = webAuthnSig,base64(publicBy),"FIDO2"
	publicType, everSig, publicBy, err := DecodeEverIdSig(txSig)
	if err != nil {
		return nil, err
	}
	switch publicType {
	case schema.FIDOPublicType:
		cred := webauthn.Credential{}
		if err = json.Unmarshal(publicBy, &cred); err != nil {
			log.Error("json.Unmarshal(public, &cred)", "err", err)
			return nil, err
		}
		_, err = verifyFidoAuthnSig(everSig, hexutil.Encode(hash), accID, GenUserId(accID, chainID), cred)
		return publicBy, err
	case schema.EVMPublicType:
		signer := common.BytesToAddress(crypto.Keccak256(publicBy[1:])[12:]).String()
		acctype, accID, err := IDCheck(signer)
		if err != nil {
			return nil, err
		}
		_, err = Verify(acctype, accID, everSig, hash, chainID)
		return publicBy, err
	case schema.ArPublicType:
		owner := utils.Base64Encode(publicBy)
		signer, err := utils.OwnerToAddress(owner)
		if err != nil {
			return nil, err
		}
		acctype, accID, err := IDCheck(signer)
		if err != nil {
			return nil, err
		}

		_, err = Verify(acctype, accID, everSig, hash, chainID)
		return publicBy, err
	default:
		return nil, schema.ERR_INVALID_ACCOUNT_TYPE
	}
}

func DecodeEverIdSig(txSig string) (publicType, everSig string, public []byte, err error) {
	// txSig = sig + "," + public + "," + publicType
	// eccSig = eccSig,base64(publicBy),"ECDSA"
	// arSig = arSig,base64(publicBy),"RSA"
	// fidoSig = webAuthnSig,base64(publicBy),"FIDO2"
	ss := strings.SplitN(txSig, ",", 3)
	if len(ss) != 3 {
		err = fmt.Errorf("invalid length of txSig:%s", txSig)
		return
	}
	everSig = ss[0]
	publicStr := ss[1]
	publicType = ss[2]
	public, err = utils.Base64Decode(publicStr)

	if publicType == schema.ArPublicType { // arweave sig = sig+","+public
		everSig = everSig + "," + publicStr
	}
	return
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

func DecodeArSig(sig string) (s []byte, pub *rsa.PublicKey, addr string, err error) {
	s, owner, err := SplitArSig(sig)
	if err != nil {
		return
	}

	addr, err = utils.OwnerToAddress(owner)
	if err != nil {
		return
	}

	pub, err = utils.OwnerToPubKey(owner)
	if err != nil {
		return
	}
	return
}

func SplitArSig(withoutPubSig string) (sig []byte, owner string, err error) {
	ss := strings.Split(withoutPubSig, ",")
	if len(ss) != 2 {
		err = fmt.Errorf("invalid length of withoutPubSig:%s", withoutPubSig)
		return
	}
	sig, err = utils.Base64Decode(ss[0])
	if err != nil {
		return
	}
	owner = ss[1]
	return
}

func verifyFidoAuthnSig(sig, hexHash string, accName, userId string, public webauthn.Credential) (*webauthn.Credential, error) {
	sigBy, err := decodeBase64(sig)
	if err != nil {
		return nil, err
	}
	authn := schema.Authn{}
	if err = json.Unmarshal(sigBy, &authn); err != nil {
		return nil, err
	}
	// 解析 car
	rawId, err := decodeBase64(authn.RawId)
	if err != nil {
		return nil, err
	}
	ClientDataJSON, err := decodeBase64(authn.ClientDataJSON)
	if err != nil {
		return nil, err
	}
	AuthenticatorData, err := decodeBase64(authn.AuthenticatorData)
	if err != nil {
		return nil, err
	}
	Signature, err := decodeBase64(authn.Signature)
	if err != nil {
		return nil, err
	}
	UserHandle, err := decodeBase64(authn.UserHandle)
	if err != nil {
		return nil, err
	}
	UserHandle = protocol.URLEncodedBase64{} // not need verify userHandle userId
	car := protocol.CredentialAssertionResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   authn.Id,
				Type: "public-key",
			},
			RawID:                   rawId,
			ClientExtensionResults:  nil,
			AuthenticatorAttachment: "platform",
		},
		AssertionResponse: protocol.AuthenticatorAssertionResponse{
			AuthenticatorResponse: protocol.AuthenticatorResponse{
				ClientDataJSON: ClientDataJSON,
			},
			AuthenticatorData: AuthenticatorData,
			Signature:         Signature,
			UserHandle:        UserHandle,
		},
	}
	pca, err := car.Parse()
	if err != nil {
		return nil, err
	}

	// new user
	user := &schema.User{userId, accName, public}
	session := webauthn.SessionData{
		Challenge:            utils.Base64Encode([]byte(hexHash)),
		UserID:               user.WebAuthnID(),
		AllowedCredentialIDs: nil,
		Expires:              time.Time{},
		UserVerification:     protocol.VerificationRequired,
		Extensions:           nil,
	}

	webAuthn, err := GetWebAuthn(pca.Response.AuthenticatorData.RPIDHash)
	if err != nil {
		log.Error("Auth GetWebAuthn", "err", err)
		return nil, err
	}
	credential, err := webAuthn.ValidateLogin(user, session, pca)
	if err != nil {
		return nil, err
	}
	return credential, nil
}

func GenEverId(email string) string {
	e := strings.ToLower(email)
	hash := sha256.Sum256([]byte(e))
	idBytes := hash[:]
	sum := checkSum(idBytes)
	fullBytes := append(idBytes, sum...)
	id := schema.EverIdPrefix + hex.EncodeToString(fullBytes) // ever + hash + checkSum
	return id
}

func IsEmailAddress(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
