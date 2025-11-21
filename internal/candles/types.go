package candles

import "time"

type OHLC struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Timestamp int64
	Interval  time.Duration
	Exchange  string
}
