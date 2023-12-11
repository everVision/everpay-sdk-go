package schema

type ChEverTx struct {
	ID int64 `gorm:"primaryKey" json:"id"`

	EverHash       string `gorm:"type:varchar(66);index:idx1" json:"everHash"`
	TokenTag       string `gorm:"index:idx2" json:"tokenTag"`
	TokenSymbol    string `json:"tokenSymbol"`
	Action         string `json:"action"`
	From           string `json:"from"`
	To             string `json:"to"`
	Amount         string `json:"amount"`
	Fee            string `json:"fee"`
	FeeRecipient   string `json:"feeRecipient"`
	Nonce          int64  `json:"nonce"`
	TokenID        string `json:"tokenID"`
	ChainType      string `json:"chainType"`
	ChainID        string `json:"chainID"`
	Data           string `json:"data"`
	Version        string `json:"version"`
	Sig            string `json:"sig"`
	InternalStatus string `json:"internalStatus"` // if internal tx (bundle tx) execute success, return "success" then return err info
}
