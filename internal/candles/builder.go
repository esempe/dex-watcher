package candles

import (
	"context"
	"planner/internal/watcher"
	"sync"
	"time"
)

type Builder struct {
	interval   time.Duration
	inputChan  <-chan watcher.PriceRatio
	outputChan chan OHLC
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	candles    map[string]*OHLC
	mu         sync.RWMutex
}

func NewBuilder(interval time.Duration, inputChan <-chan watcher.PriceRatio) *Builder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Builder{
		interval:   interval,
		inputChan:  inputChan,
		outputChan: make(chan OHLC, 10),
		ctx:        ctx,
		cancel:     cancel,
		candles:    make(map[string]*OHLC),
	}
}

func (b *Builder) process() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			return
		case ratio := <-b.inputChan:
			b.updateCandle(ratio)
		case <-ticker.C:
			b.flushCandles()
		}
	}
}

func (b *Builder) Start() {
	b.wg.Add(1)
	go b.process()
}

func (b *Builder) updateCandle(ratio watcher.PriceRatio) {
	key := "HL/L"
	b.mu.Lock()
	defer b.mu.Unlock()

	candle, exists := b.candles[key]
	if !exists || candle.Timestamp+int64(b.interval.Seconds()) <= ratio.Timestamp {
		if exists {
			select {
			case b.outputChan <- *candle:
			case <-b.ctx.Done():
				return
			}
		}
		candle = &OHLC{
			Symbol:    "BTC",
			Open:      ratio.Ratio,
			High:      ratio.Ratio,
			Low:       ratio.Ratio,
			Close:     ratio.Ratio,
			Timestamp: ratio.Timestamp,
			Interval:  b.interval,
			Exchange:  "HL/L",
		}
		b.candles[key] = candle
	} else {
		if ratio.Ratio > candle.High {
			candle.High = ratio.Ratio
		}
		if ratio.Ratio < candle.Low {
			candle.Low = ratio.Ratio
		}
		candle.Close = ratio.Ratio
	}
}

func (b *Builder) flushCandles() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, candle := range b.candles {
		select {
		case b.outputChan <- *candle:
		case <-b.ctx.Done():
			return
		}
	}
	b.candles = make(map[string]*OHLC)
}

func (b *Builder) CandleChannel() <-chan OHLC {
	return b.outputChan
}

func (b *Builder) Stop() {
	b.cancel()
	b.wg.Wait()
	close(b.outputChan)
}
