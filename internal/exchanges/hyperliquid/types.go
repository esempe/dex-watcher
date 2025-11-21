package hyperliquid

type HyperliquidResponse struct {
	Channel string  `json:"channel"`
	Data    []Trade `json:"data"`
}

// Trade - структура одной сделки
type Trade struct {
	Coin  string   `json:"coin"`
	Hash  string   `json:"hash"`
	Px    string   `json:"px"`
	Side  string   `json:"side"`
	Sz    string   `json:"sz"`
	Tid   int64    `json:"tid"`
	Time  int64    `json:"time"`
	Users []string `json:"users"`
}
