package schema

type TargetChain struct {
	ChainID   string `json:"targetChainId"`
	ChainType string `json:"targetChainType"`         // e.g: "avalanche" "arweave" "ethereum","moon"
	Decimals  int    `json:"targetDecimals"`          // e.g: 18
	TokenId   string `json:"targetTokenId,omitempty"` // target chain token address
}
