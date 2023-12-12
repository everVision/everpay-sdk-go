package schema

import (
	"github.com/everFinance/goar/utils"
	"github.com/go-webauthn/webauthn/webauthn"
)

const (
	EverIdPrefix = "eid"
	EverIdLength = 71

	EVMPublicType  = "ECDSA"
	ArPublicType   = "RSA"
	FIDOPublicType = "FIDO2"

	AccountTypeEVM    = "ethereum"
	AccountTypeAR     = "arweave"
	AccountTypeEverId = "eid"

	// rpId
	LocalhostRpId = "localhost"
	EverpayRpId   = "everpay.io"

	// rp origins
	LocalhostOrg      = "http://localhost:8080"
	EverpayOrg        = "https://app.everpay.io"
	EverpayDevOrg     = "https://app-dev.everpay.io"
	BetaDevEverpayOrg = "https://beta-dev.everpay.io"
	BetaEverpayOrg    = "https://beta.everpay.io"
)

type Authn struct {
	Id                string `json:"id"`
	RawId             string `json:"rawId"`
	ClientDataJSON    string `json:"clientDataJSON"`
	AuthenticatorData string `json:"authenticatorData"`
	Signature         string `json:"signature"`
	UserHandle        string `json:"userHandle"`
}

type User struct {
	Id     string
	Name   string
	Public webauthn.Credential // key: publicType，val: credential
}

func (u *User) WebAuthnID() []byte {
	id, _ := utils.Base64Decode(u.Id)
	return id
}

func (u *User) WebAuthnName() string {
	return u.Name
}

func (u *User) WebAuthnDisplayName() string {
	return u.Name
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{u.Public}
}

func (u *User) WebAuthnIcon() string {
	return ""
}
