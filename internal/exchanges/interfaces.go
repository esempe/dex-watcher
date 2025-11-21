package exchanges

type ExchangeClient interface {
	Connect() error
	Subscribe(symbol string) error
	PriceChannel() <-chan float64
	Close() error
	IsConnected() bool
	Name() string
}

type PriceUpdate struct {
	Symbol    string
	Price     float64
	Timestamp int64
	Exchange  string
}

type OCHL struct {
}
