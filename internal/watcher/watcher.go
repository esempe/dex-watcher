package watcher

import (
	"context"
	"planner/internal/exchanges"
	"planner/internal/exchanges/hyperliquid"
	"planner/internal/exchanges/lighter"
	"sync"
)

type Watcher struct {
	hClient    *hyperliquid.HyperliquidClient
	lClient    *lighter.LighterClient
	outputChan chan PriceRatio
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
	lastHL     *exchanges.PriceUpdate
	lastL      *exchanges.PriceUpdate
}

func New(ticker string) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	hClient := hyperliquid.NewHyperliquidClient(ticker)
	lClient := lighter.NewLighterClient(ticker)
	return &Watcher{
		hClient:    hClient,
		lClient:    lClient,
		ctx:        ctx,
		cancel:     cancel,
		outputChan: make(chan PriceRatio, 20),
	}
}

func (w *Watcher) StartRatioCalculator() error {
	if err := w.hClient.Connect(); err != nil {
		return err
	}
	if err := w.lClient.Connect(); err != nil {
		return err
	}
	if err := w.hClient.Subscribe(); err != nil {
		return err
	}
	if err := w.lClient.Subscribe(); err != nil {
		return err
	}

	w.wg.Add(2)
	go w.readHLChannel()
	go w.readLChannel()
	return nil
}

func (w *Watcher) readHLChannel() {
	defer w.wg.Done()
	hlChan := w.hClient.PriceChannel()
	for {
		select {
		case <-w.ctx.Done():
			return
		case price, ok := <-hlChan:
			if !ok {
				return
			}
			w.mu.Lock()
			w.lastHL = &price
			var lastLPrice float64
			hasLastL := w.lastL != nil
			if hasLastL {
				lastLPrice = w.lastL.Price
			}
			w.mu.Unlock()

			if hasLastL {
				ratio := price.Price / lastLPrice
				select {
				case w.outputChan <- PriceRatio{
					Ratio:     ratio,
					HLPrice:   price.Price,
					LPrice:    lastLPrice,
					Timestamp: price.Timestamp,
				}:
				case <-w.ctx.Done():
					return
				}
			}
		}
	}
}

func (w *Watcher) readLChannel() {
	defer w.wg.Done()
	lChan := w.lClient.PriceChannel()
	for {
		select {
		case <-w.ctx.Done():
			return
		case price, ok := <-lChan:
			if !ok {
				return
			}
			w.mu.Lock()
			w.lastL = &price
			var lastHLPrice float64
			hasLastHL := w.lastHL != nil
			if hasLastHL {
				lastHLPrice = w.lastHL.Price
			}

			w.mu.Unlock()

			if hasLastHL {
				ratio := lastHLPrice / price.Price
				select {
				case w.outputChan <- PriceRatio{
					Ratio:     ratio,
					HLPrice:   lastHLPrice,
					LPrice:    price.Price,
					Timestamp: price.Timestamp,
				}:
				case <-w.ctx.Done():
					return
				}
			}
		}
	}
}

func (w *Watcher) RatioChannel() <-chan PriceRatio {
	return w.outputChan
}

func (w *Watcher) Stop() {
	w.cancel()
	w.wg.Wait()
	close(w.outputChan)
}
