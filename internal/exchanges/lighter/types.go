package lighter

type LighterResponse struct {
	Channel           string  `json:"channel"`
	LiquidationTrades []Trade `json:"liquidation_trades"`
	Nonce             int64   `json:"nonce"`
	Trades            []Trade `json:"trades"`
	Type              string  `json:"type"`
}

type Trade struct {
	AskAccountID                     int64  `json:"ask_account_id"`
	AskClientID                      int64  `json:"ask_client_id"`
	AskID                            int64  `json:"ask_id"`
	BidAccountID                     int64  `json:"bid_account_id"`
	BidClientID                      int64  `json:"bid_client_id"`
	BidID                            int64  `json:"bid_id"`
	BlockHeight                      int64  `json:"block_height"`
	IsMakerAsk                       bool   `json:"is_maker_ask"`
	MakerEntryQuoteBefore            string `json:"maker_entry_quote_before"`
	MakerFee                         int    `json:"maker_fee"`
	MakerInitialMarginFractionBefore int    `json:"maker_initial_margin_fraction_before"`
	MakerPositionSizeBefore          string `json:"maker_position_size_before"`
	MarketID                         uint8  `json:"market_id"`
	Price                            string `json:"price"`
	Size                             string `json:"size"`
	TakerEntryQuoteBefore            string `json:"taker_entry_quote_before"`
	TakerInitialMarginFractionBefore int    `json:"taker_initial_margin_fraction_before"`
	TakerPositionSizeBefore          string `json:"taker_position_size_before"`
	Timestamp                        int64  `json:"timestamp"`
	TradeID                          int64  `json:"trade_id"`
	TxHash                           string `json:"tx_hash"`
	Type                             string `json:"type"`
	USDAmount                        string `json:"usd_amount"`
}

var TickerToID = map[string]uint8{
	"BTC":  1,
	"ETH":  0,
	"HYPE": 24,
}

var IDToTicker = map[uint8]string{
	1:  "BTC",
	0:  "ETH",
	24: "HYPE",
}
