package lighter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"planner/internal/exchanges"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

type LighterClient struct {
	wsUrl       string
	conn        *websocket.Conn
	isConnected bool
	marketId    uint8
	priceChan   chan exchanges.PriceUpdate
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func NewLighterClient(symbol string) *LighterClient {
	ctx, cancel := context.WithCancel(context.Background())
	marketId := tickerToMarketId(symbol)

	return &LighterClient{
		wsUrl:     "wss://mainnet.zklighter.elliot.ai/stream",
		marketId:  marketId,
		priceChan: make(chan exchanges.PriceUpdate, 10),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (c *LighterClient) Subscribe() error {
	c.mu.RLock()
	if !c.isConnected || c.conn == nil {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	c.mu.RUnlock()

	subscription := map[string]any{
		"type":    "subscribe",
		"channel": fmt.Sprintf("trade/%d", c.marketId),
	}

	if err := c.conn.WriteJSON(subscription); err != nil {
		return err
	}

	log.Printf("Subscribed to %s on Lighter", marketIdToTicker(c.marketId))
	return nil
}

func (c *LighterClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isConnected {
		return nil
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(c.wsUrl, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	c.isConnected = true
	go c.readMessages()

	log.Printf("Connected to Hyperliquid WebSocket for %s", marketIdToTicker(c.marketId))

	return nil
}

func (c *LighterClient) readMessages() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}
			var response LighterResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Printf("Error parsing message: %v", err)
				continue
			}
			if priceData := c.extractPrice(response); priceData.Price > 0 {
				select {
				case c.priceChan <- priceData:
				case <-c.ctx.Done():
					return
				}
			}
		}
	}
}

func (c *LighterClient) extractPrice(response LighterResponse) exchanges.PriceUpdate {

	if len(response.Trades) == 0 {
		return exchanges.PriceUpdate{}
	}

	price, err := strconv.ParseFloat(response.Trades[0].Price, 64) //
	if err != nil {
		log.Printf("Error parsing price: %v", err)
		return exchanges.PriceUpdate{}
	}

	return exchanges.PriceUpdate{
		Symbol:    "btc",
		Price:     price,
		Timestamp: response.Trades[0].Timestamp,
		Exchange:  "Lighter",
	}
}

func (c *LighterClient) PriceChannel() <-chan exchanges.PriceUpdate {
	return c.priceChan
}

func (c *LighterClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	c.cancel()

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return err
		}
	}

	c.isConnected = false
	close(c.priceChan)

	log.Println("Closed Hyperliquid WebSocket connection")
	return nil
}

func (c *LighterClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

func (c *LighterClient) Name() string {
	return "Hyperliquid"
}

func tickerToMarketId(ticker string) uint8 {
	if id, ok := TickerToID[ticker]; ok {
		return uint8(id)
	}
	return 0
}

func marketIdToTicker(id uint8) string {

	if ticker, ok := IDToTicker[id]; ok {
		return ticker
	}
	return ""
}

var ErrNotConnected = errors.New("not connected to exchange")
