package schema

type RespErr struct {
	Err string `json:"error"`
}

func (r RespErr) Error() string {
	return r.Err
}

type WithdrawTxResponse struct {
	EverHash    string
	Token       string
	Status      string
	WithdrawFee string
	WithdrawTx  string
	RefundTx    string
	Error       string
}

type TokenInfo struct {
	Tag string `json:"tag"`
	// tokMeta
	ID                 string                 `json:"id"` // On Native-Chain tokenId; Special AR token: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xcc9141efa8c20c7df0778748255b1487957811be"
	Symbol             string                 `json:"symbol"`
	Decimals           int                    `json:"decimals"`  // On everPay decimals
	ChainType          string                 `json:"chainType"` // On everPay chainType; Special AR token: "arweave,ethereum"
	ChainID            string                 `json:"chainID"`   // On everPay chainId; Special AR token: "0,1"(mainnet) or "0,42"(testnet)
	Display            bool                   `json:"display"`
	Type               string                 `json:"type"`
	TotalSupply        string                 `json:"totalSupply"`
	BurnFees           map[string]string      `json:"burnFees"` // key: targetChainType, val: fee
	TransferFee        string                 `json:"transferFee"`
	BundleFee          string                 `json:"bundleFee"`
	HolderNum          int64                  `json:"holderNum"`
	CrossChainInfoList map[string]TargetChain `json:"crossChainInfoList"` // key: targetChainType
	TNS102Extra        *Tns102Extra           `json:"tns102Extra"`
}

type Tns102Extra struct {
	Owner          string `json:"owner"`
	PauseBlackList bool   `json:"pauseBlackList"` // default false
	PauseWhiteList bool   `json:"pauseWhiteList"` // default false
	Pause          bool   `json:"pause"`          // pause token default fals
}

type Info struct {
	IsSynced       bool              `json:"isSynced"`
	IsClosed       bool              `json:"isClosed"`
	Owner          string            `json:"owner"`
	SetActionOwner string            `json:"setActionOwner"`
	EthChainID     string            `json:"ethChainID"`
	FeeRecipient   string            `json:"feeRecipient"`
	EthLocker      string            `json:"ethLocker"`
	ArLocker       string            `json:"arLocker"`
	Lockers        map[string]string `json:"lockers"`
	TokenList      []TokenInfo       `json:"tokenList"`
}

type TuringInfo struct {
	BalanceMerkleRoot BalanceTreeHash
	RollupCurrentID   string `json:"rollupCurrentID"`
	RollupParentID    string `json:"rollupParentID"`
	PendingTxNum      int64  `json:"pendingTxNum"`
	CurRollupTxNum    int64  `json:"curRollupTxNum"`
	RollupAddr        string `json:"rollupAddr"`
}

type LimitIp struct {
	Limit bool `json:"limit"`
}

type Balance struct {
	Tag      string `json:"tag"`
	Amount   string `json:"amount"`
	Decimals int    `json:"decimals"`
}

type AccBalance struct {
	AccId   string  `json:"accid"`
	Balance Balance `json:"balance"`
}

type AccBalances struct {
	AccId    string    `json:"accid"`
	Balances []Balance `json:"balances"`
}

type Txs struct {
	Txs         []TxResponse `json:"txs"`
	HasNextPage bool         `json:"hasNextPage"`
}

type AccTxs struct {
	AccId string `json:"accid"`
	Txs
}

type Tx struct {
	Tx *TxResponse `json:"tx"`
}

type RespStatus struct {
	Status string `json:"status"`
}

type PendingTxs struct {
	HasNextPage bool       `json:"hasNextPage"` // true means can get more
	Txs         []ChEverTx `json:"txs"`
}

type Fee struct {
	Fee TokenFee `json:"fee"`
}

type Fees struct {
	Fees []TokenFee `json:"fees"`
}

type RespAcc struct {
	Id           string            `json:"id"`
	Type         string            `json:"type"`
	PublicType   map[string]string `json:"publicType"`   // key: publicId, val: publicType
	PublicValues map[string]string `json:"publicValues"` // key: publicId, val: public base64encode
}

type RespRegister struct {
	Sig       string `json:"sig"`
	Timestamp int64  `json:"timestamp"`
}

type TxOpts struct {
	Address       string
	TokenTag      string
	Action        string
	WithoutAction string
}
